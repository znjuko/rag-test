#!/usr/bin/env python3

"""
Multi-Format Document Processing with Docling
==============================================

This script demonstrates Docling's ability to handle multiple
document formats with a unified API.

Supported formats:
- PDF (.pdf)
- Word (.docx, .doc)
- PowerPoint (.pptx, .ppt)
- Excel (.xlsx, .xls)
- HTML (.html, .htm)
- Images (.png, .jpg)
- And more...

"""

import os
import sys
from pathlib import Path
from docling.document_converter import DocumentConverter

def process_document(file_path: str, output_path: str, converter: DocumentConverter) -> dict:
    """Process a single document and return metadata."""
    try:
        print(f"\n📄 Processing: {Path(file_path).name}")

        result = converter.convert(file_path)
        markdown = result.document.export_to_markdown()

        doc_info = {
            'file': Path(file_path).name,
            'format': Path(file_path).suffix,
            'status': 'Success',
            'markdown_length': len(markdown),
            'preview': markdown[:200].replace('\n', ' ')
        }

        output_file = f"{output_path}/{Path(file_path).stem}.md"
        with open(output_file, 'w', encoding='utf-8') as f:
            f.write(markdown)

        doc_info['output_file'] = output_file

        print(f"   ✓ Converted successfully")
        print(f"   ✓ Output: {output_file}")

        return doc_info

    except Exception as e:
        print(f"   ✗ Error: {e}")
        return {
            'file': Path(file_path).name,
            'format': Path(file_path).suffix,
            'status': 'Failed',
            'error': str(e)
        }

def main():
    print("=" * 60)
    print("Multi-Format Document Processing with Docling")
    print("=" * 60)

    read_path = sys.argv[1]
    output_path = sys.argv[2]

    # Initialize converter once (reusable)
    converter = DocumentConverter()

    result = process_document(read_path, output_path, converter)
        
    # Summary
    print("\n" + "=" * 60)
    print("CONVERSION SUMMARY")
    print("=" * 60)

    status_icon = "✓" if result['status'] == 'Success' else "✗"
    print(f"{status_icon} {result['file']} ({result['format']})")
    if result['status'] == 'Success':
        print(f"   Length: {result['markdown_length']} chars")
        print(f"   Preview: {result['preview']}...")
    else:
        print(f"   Error: {result.get('error', 'Unknown')}")
    print()

if __name__ == "__main__":
    main()