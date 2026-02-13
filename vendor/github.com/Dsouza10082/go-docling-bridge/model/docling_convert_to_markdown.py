#!/usr/bin/env python3
import sys
import json
import os

def convert_to_markdown(file_path):
    """Convert document to markdown using Docling"""
    try:
        from docling.document_converter import DocumentConverter
        
        converter = DocumentConverter()
        result = converter.convert(file_path)
        markdown = result.document.export_to_markdown()
        
        return {
            "status": "success",
            "markdown": markdown,
            "file": os.path.basename(file_path)
        }
    except ImportError as e:
        return {
            "status": "failed",
            "error": f"Docling not installed. Run: pip install docling. Error: {str(e)}",
            "file": os.path.basename(file_path)
        }
    except Exception as e:
        return {
            "status": "failed",
            "error": str(e),
            "file": os.path.basename(file_path)
        }

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("JSON_RESULT_START")
        print(json.dumps({"status": "failed", "error": "No file path provided"}))
        print("JSON_RESULT_END")
        sys.exit(1)
    
    file_path = sys.argv[1]
    result = convert_to_markdown(file_path)
    
    print("JSON_RESULT_START")
    print(json.dumps(result, ensure_ascii=False))
    print("JSON_RESULT_END")