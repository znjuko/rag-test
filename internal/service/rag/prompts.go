package rag

import "fmt"

const (
	analysisSystemPrompt = `Ты — модуль анализа запроса для RAG. Твоя задача: решить, нужно ли уточнение.
Не отвечай на вопрос пользователя. Не придумывай факты.
Ответ должен быть ТОЛЬКО валидным JSON без комментариев, без пояснений и без markdown.
Строго следуй схеме:
{
  "need_clarification": boolean,
  "clarifying_question": string | null,
  "missing_slots": string[],
  "assumptions": string[]
}`

	rewriteSystemPrompt = `Ты — модуль переписывания запросов для поиска (RAG retrieval).
Не отвечай на вопрос пользователя. Не добавляй факты.
Сгенерируй несколько поисковых запросов, сохраняя смысл.
Ответ должен быть ТОЛЬКО валидным JSON без комментариев, без пояснений и без markdown.
Строго следуй схеме:
{
  "queries": string[]
}`

	answerSystemPrompt = `Ты — RAG-ассистент. Отвечай ТОЛЬКО на основе предоставленных фрагментов.
	Не добавляй знания извне. Не делай догадок.

	Ответ должен быть ТОЛЬКО валидным JSON (без комментариев, без пояснений и без markdown),
	строго по схеме:
{
  "text": "...",
  "citations_used": ["C1","C2",...],
  "citations": [
    {
      "id": "C1",
      "data_source": "...",
      "quote": "..."
    }
  ]
}

Если в фрагментах нет ответа, то верни JSON по той же схеме, где:
- "text": "Не знаю на основе предоставленных источников."
- "citations_used": []
- "citations": []

	Требования к ответу:
	1) Дай краткий структурированный ответ.
	2) Для каждого ключевого утверждения укажи ссылку-метку вида [C1], [C2]...
	3) В конце выведи раздел "Источники" со списком C1..Cn, где для каждой метки:
	   - data_source
	   - цитируемый фрагмент (короткий текст из источника)`

	answerRewriteSystemPrompt = `Ты — модуль исправления ответа. Исправь ответ, учитывая замечания валидатора.
	Не добавляй знания извне. Не делай догадок.

	Ответ должен быть ТОЛЬКО валидным JSON (без комментариев, без пояснений и без markdown),
	строго по схеме:
	{
	  "text": "...",
	  "citations_used": ["C1","C2",...],
	  "citations": [
	    {
	      "id": "C1",
	      "data_source": "...",
	      "quote": "..."
	    }
	  ]
	}

	Если после исправлений всё равно нет подтверждений в фрагментах, то верни JSON по той же схеме, где:
	- "text": "Не знаю на основе предоставленных источников."
	- "citations_used": []
	- "citations": []`

	validationSystemPrompt = `Ты — валидатор. Проверь, что утверждения в ответе подтверждаются указанными источниками, если есть обобщения - игнорируй такие утверждения и не валидируй их.
Общие данные валидировать не валидируй, валидируй только то, что имеет четкую конкретику, четкое отношение к тому или иному объекту данных.
Если есть утверждения без опоры — верни ok=false.
Ответ должен быть ТОЛЬКО валидным JSON без комментариев, без пояснений и без markdown.
Строго следуй схеме:
{
  "ok": boolean,
  "unsupported_claims": string[],
  "notes": string
}`
)

func buildClarificationUserPrompt(question, dialogContext string) string {
	return fmt.Sprintf(`Вопрос пользователя: %s

Контекст диалога (может быть пустым):
%s

Правила:
- need_clarification=true, если без уточнения высок риск неверного ответа
  (двусмысленность, не указан объект/продукт/версия/регион/период).
- Если можно безопасно искать без уточнения — need_clarification=false.

JSON-схема (строго, только JSON, без текста вокруг):
{
  "need_clarification": boolean,
  "clarifying_question": string | null,
  "missing_slots": string[],
  "assumptions": string[]
}

Слоты (примеры): ["product", "product_version", "region", "time_period", "user_role", "entity"]`, question, dialogContext)
}

func buildRewriteUserPrompt(question, dialogContext, analysisJSON string) string {
	return fmt.Sprintf(`Вопрос пользователя: %s

Контекст диалога (может быть пустым):
%s

Анализ запроса (JSON):
%s

JSON-схема (строго, только JSON, без текста вокруг):
{
  "queries": string[]
}`, question, dialogContext, analysisJSON)
}

func buildAnswerUserPrompt(question, dialogContext, chunks string) string {
	return fmt.Sprintf(`Вопрос: %s
Контекст диалога: %s

Фрагменты (каждый с id и метаданными):
%s

Ответ верни в формате (строго, только JSON, без текста вокруг):
{
  "text": "...",
  "citations_used": ["C1","C2",...],
  "citations": [
    {
      "id": "C1",
      "data_source": "...",
      "quote": "..."
    }
  ]
}
Если ты не можешь сослаться минимум на один фрагмент, который явно подтверждает ответ,
то верни JSON по той же схеме, где:
- "text": "%s"
- "citations_used": []
- "citations": []`, question, dialogContext, chunks, unknownAnswer)
}

func buildAnswerRewriteUserPrompt(question, dialogContext, chunks, answerText, validationFeedback string) string {
	return fmt.Sprintf(`Вопрос: %s
Контекст диалога: %s

Предыдущий ответ:
%s

Замечания валидатора:
%s

Фрагменты (каждый с id и метаданными):
%s

Ответ верни в формате (строго, только JSON, без текста вокруг):
{
  "text": "...",
  "citations_used": ["C1","C2",...],
  "citations": [
    {
      "id": "C1",
      "data_source": "...",
      "quote": "..."
    }
  ]
}

Если не можешь сослаться минимум на один фрагмент, который явно подтверждает ответ,
то верни JSON по той же схеме, где:
- "text": "%s"
- "citations_used": []
- "citations": []`, question, dialogContext, answerText, validationFeedback, chunks, unknownAnswer)
}

func buildValidationUserPrompt(question, answerText, chunks string) string {
	return fmt.Sprintf(`Вопрос: %s

Ответ:
%s

Источники с текстом:
%s

JSON (строго, только JSON, без текста вокруг):
{
  "ok": boolean,
  "unsupported_claims": string[],
  "notes": string
}`, question, answerText, chunks)
}
