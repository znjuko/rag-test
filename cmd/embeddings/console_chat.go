package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	openairepo "rag-test/internal/repository/openai"
	"rag-test/internal/service/rag"
)

const (
	chatPrompt       = "you> "
	dividerLine      = "------------------------------------------------------------"
	maxPreviewRunes  = 220
	defaultScanLimit = 1024 * 1024
)

type chatState struct {
	history []rag.DialogMessage
	topK    int
}

func runConsoleChat(ctx context.Context, ragSvc *rag.Service) error {
	if ragSvc == nil {
		return fmt.Errorf("rag service is nil")
	}

	state := chatState{
		history: make([]rag.DialogMessage, 0, 32),
		topK:    5,
	}

	printChatIntro(os.Stdout)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), defaultScanLimit)

	for {
		fmt.Fprint(os.Stdout, chatPrompt)
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		handled, stop := handleCommand(os.Stdout, &state, line)
		if handled {
			if stop {
				return nil
			}
			continue
		}

		n := time.Now()
		resp, err := ragSvc.Answer(ctx, rag.Request{
			Question: line,
			History:  state.history,
			TopK:     state.topK,
		})
		if err != nil {
			fmt.Fprintf(os.Stdout, "Ошибка: %v\n", err)
			continue
		}

		renderResponse(os.Stdout, line, resp, n)
		appendHistory(&state, line, resp)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func printChatIntro(out io.Writer) {
	fmt.Fprintln(out, dividerLine)
	fmt.Fprintln(out, "RAG чат запущен. Введите вопрос и нажмите Enter.")
	fmt.Fprintln(out, "Команды: /help, /exit, /quit, /clear, /topk N")
	fmt.Fprintln(out, dividerLine)
}

func handleCommand(out io.Writer, state *chatState, line string) (bool, bool) {
	if !strings.HasPrefix(line, "/") {
		return false, false
	}

	fields := strings.Fields(line)
	if len(fields) == 0 {
		return true, false
	}

	switch strings.ToLower(fields[0]) {
	case "/exit", "/quit":
		fmt.Fprintln(out, "Завершаю чат.")
		return true, true
	case "/clear":
		state.history = nil
		fmt.Fprintln(out, "История очищена.")
		return true, false
	case "/topk":
		if len(fields) < 2 {
			fmt.Fprintf(out, "Текущий topK: %d\n", state.topK)
			return true, false
		}
		value, err := strconv.Atoi(fields[1])
		if err != nil || value < 0 {
			fmt.Fprintln(out, "Неверное значение. Пример: /topk 8")
			return true, false
		}
		state.topK = value
		fmt.Fprintf(out, "topK установлен: %d\n", state.topK)
		return true, false
	case "/help":
		fmt.Fprintln(out, "Доступные команды:")
		fmt.Fprintln(out, "- /help  показать справку")
		fmt.Fprintln(out, "- /exit  выйти из чата")
		fmt.Fprintln(out, "- /quit  выйти из чата")
		fmt.Fprintln(out, "- /clear очистить историю")
		fmt.Fprintln(out, "- /topk N задать число контекстных чанков (0 = по умолчанию)")
		return true, false
	default:
		fmt.Fprintln(out, "Неизвестная команда. Используйте /help.")
		return true, false
	}
}

func appendHistory(state *chatState, question string, resp *rag.Response) {
	if state == nil {
		return
	}

	state.history = append(state.history, rag.DialogMessage{
		Role:    openairepo.RoleUser,
		Content: question,
	})

	answer := assistantMessage(resp)
	if answer == "" {
		return
	}

	state.history = append(state.history, rag.DialogMessage{
		Role:    openairepo.RoleAssistant,
		Content: answer,
	})
}

func assistantMessage(resp *rag.Response) string {
	if resp == nil {
		return ""
	}
	if resp.NeedClarification {
		if strings.TrimSpace(resp.ClarifyingQuestion) != "" {
			return strings.TrimSpace(resp.ClarifyingQuestion)
		}
		return "Нужно уточнение по запросу."
	}
	return strings.TrimSpace(resp.Answer)
}

func renderResponse(out io.Writer, question string, resp *rag.Response, n time.Time) {
	if resp == nil {
		fmt.Fprintln(out, "Пустой ответ от сервиса.")
		return
	}

	fmt.Fprintln(out, dividerLine)
	fmt.Fprintf(out, "Запрос: %s\n", question)
	fmt.Fprintln(out, dividerLine)
	fmt.Fprintf(out, "Время ответа: %s\n", time.Since(n).String())
	fmt.Fprintln(out, dividerLine)

	if resp.NeedClarification {
		fmt.Fprintln(out, "Статус: требуется уточнение")
		if strings.TrimSpace(resp.ClarifyingQuestion) != "" {
			fmt.Fprintf(out, "Уточняющий вопрос: %s\n", strings.TrimSpace(resp.ClarifyingQuestion))
		}
		printList(out, "Недостающие параметры", resp.MissingSlots)
		printList(out, "Предположения", resp.Assumptions)
		printList(out, "Предложенные запросы", resp.SuggestedQueries)
		fmt.Fprintln(out, dividerLine)
		return
	}

	fmt.Fprintln(out, "Статус: ответ готов")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Ответ:")
	fmt.Fprintln(out, strings.TrimSpace(resp.Answer))

	printList(out, "Цитаты использованы", resp.CitationsUsed)
	printCitations(out, resp.Citations)
	printChunks(out, resp.Chunks)
	printValidation(out, resp.Validation)

	fmt.Fprintln(out, dividerLine)
}

func printList(out io.Writer, title string, items []string) {
	clean := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			clean = append(clean, item)
		}
	}
	if len(clean) == 0 {
		return
	}

	fmt.Fprintf(out, "%s:\n", title)
	for _, item := range clean {
		fmt.Fprintf(out, "- %s\n", item)
	}
}

func printCitations(out io.Writer, citations []rag.Citation) {
	if len(citations) == 0 {
		return
	}

	fmt.Fprintln(out, "Цитаты:")
	for _, c := range citations {
		quote := truncate(singleLine(c.Quote), maxPreviewRunes)
		source := strings.TrimSpace(c.DataSource)
		if source == "" {
			source = "unknown"
		}
		fmt.Fprintf(out, "- [%s] source=%s | %s\n", c.ID, source, quote)
	}
}

func printChunks(out io.Writer, chunks []rag.Chunk) {
	if len(chunks) == 0 {
		return
	}

	fmt.Fprintln(out, "Контекстные чанки:")
	for _, c := range chunks {
		text := truncate(singleLine(c.Text), maxPreviewRunes)
		source := strings.TrimSpace(c.DataSource)
		if source == "" {
			source = "unknown"
		}
		fmt.Fprintf(out, "- [%s] source=%s | %s\n", c.ID, source, text)
	}
}

func printValidation(out io.Writer, validation rag.ValidationResult) {
	status := "FAIL"
	if validation.OK {
		status = "OK"
	}
	fmt.Fprintf(out, "Валидация: %s\n", status)
	printList(out, "Неподдержанные утверждения", validation.UnsupportedClaims)
	if strings.TrimSpace(validation.Notes) != "" {
		fmt.Fprintf(out, "Примечания: %s\n", strings.TrimSpace(validation.Notes))
	}
}

func singleLine(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.TrimSpace(text)
	fields := strings.Fields(text)
	return strings.Join(fields, " ")
}

func truncate(text string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-3]) + "..."
}
