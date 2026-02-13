#!/usr/bin/env python3
"""
Docling Convert One File to Markdown
Converts a single document file to Markdown format using Docling.

Usage:
    python docling_convert_one_file_to_markdown.py <file_path>

Output:
    JSON result wrapped in markers for Go parsing:
    JSON_RESULT_START
    {"status": "success", "markdown": "...", "file": "..."}
    JSON_RESULT_END
"""

import sys
import json
import os


def convert_to_markdown(file_path: str) -> dict:
    """
    Convert a document to Markdown using Docling.
    
    Args:
        file_path: Path to the document file
        
    Returns:
        Dictionary with status, markdown content, and file name
    """
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
            "file": os.path.basename(file_path) if file_path else "",
            "markdown": ""
        }
        
    except FileNotFoundError as e:
        return {
            "status": "failed",
            "error": f"File not found: {str(e)}",
            "file": os.path.basename(file_path) if file_path else "",
            "markdown": ""
        }
        
    except Exception as e:
        return {
            "status": "failed",
            "error": str(e),
            "file": os.path.basename(file_path) if file_path else "",
            "markdown": ""
        }


def main():
    """Main entry point for CLI usage."""
    if len(sys.argv) < 2:
        result = {
            "status": "failed",
            "error": "No file path provided. Usage: python script.py <file_path>",
            "file": "",
            "markdown": ""
        }
    else:
        file_path = sys.argv[1]
        result = convert_to_markdown(file_path)
    
    # Output with markers for Go parsing
    print("JSON_RESULT_START")
    print(json.dumps(result, ensure_ascii=False))
    print("JSON_RESULT_END")


if __name__ == "__main__":
    main()