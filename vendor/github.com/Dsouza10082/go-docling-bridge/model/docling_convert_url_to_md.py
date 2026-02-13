import sys
import json

try:
    from docling.document_converter import DocumentConverter
    
    input_path = sys.argv[1]
    
    converter = DocumentConverter()
    result = converter.convert(input_path)
    doc = result.document
    
    # Export to markdown
    markdown = doc.export_to_markdown()
    
    # Try to get title
    title = ""
    if hasattr(doc, 'title') and doc.title:
        title = str(doc.title)
    
    # Output as JSON
    output = {
        "markdown": markdown,
        "title": title,
        "status": "success"
    }
    print(json.dumps(output))
    
except Exception as e:
    output = {
        "markdown": "",
        "title": "",
        "status": "error",
        "error": str(e)
    }
    print(json.dumps(output))
    sys.exit(1)