#!/usr/bin/env python3
"""
Script to generate documentation README based on docs.yaml configuration
"""

import yaml
import os
import sys
from pathlib import Path

def load_config(config_path="docs.yaml"):
    """Load configuration from docs.yaml"""
    try:
        with open(config_path, 'r', encoding='utf-8') as f:
            return yaml.safe_load(f)
    except FileNotFoundError:
        print(f"❌ Configuration file {config_path} not found")
        return None
    except yaml.YAMLError as e:
        print(f"❌ Error parsing {config_path}: {e}")
        return None

def generate_header(config):
    """Generate the header section with image and title"""
    header = config.get('header', {})
    styling = config.get('styling', {})
    
    lines = []
    
    # Title
    title = header.get('title', 'Documentation')
    lines.append(f"# {title}")
    lines.append("")
    
    # Image section (if enabled)
    if styling.get('show_image', True):
        image_config = header.get('image', {})
        if image_config:
            lines.append('<div align="center">')
            lines.append(f'  <img src="{image_config.get("src", "./images/deco_gopher.png")}" '
                        f'alt="{image_config.get("alt", "Go Gopher Artist")}" '
                        f'width="{image_config.get("width", 200)}" '
                        f'height="{image_config.get("height", 200)}">')
            lines.append("  <br>")
            caption = image_config.get('caption', '')
            if caption:
                lines.append(f'  <em>{caption}</em>')
            lines.append("</div>")
            lines.append("")
    
    # Subtitle
    subtitle = header.get('subtitle', '')
    if subtitle:
        lines.append(subtitle)
        lines.append("")
    
    return lines

def generate_sections(config):
    """Generate the documentation sections"""
    sections = config.get('sections', [])
    styling = config.get('styling', {})
    
    lines = []
    
    if sections:
        lines.append("## 📚 Documentation Sections")
        lines.append("")
        
        for section in sections:
            title = section.get('title', '')
            file = section.get('file', '')
            description = section.get('description', '')
            
            if title and file:
                link = f"- [{title}](./{file})"
                if description:
                    link += f" - {description}"
                lines.append(link)
        
        lines.append("")
    
    return lines

def generate_quick_links(config):
    """Generate the quick links section"""
    quick_links = config.get('quick_links', [])
    styling = config.get('styling', {})
    
    lines = []
    
    if quick_links:
        lines.append("## 🚀 Quick Links")
        lines.append("")
        
        for link in quick_links:
            title = link.get('title', '')
            url = link.get('url', '')
            
            if title and url:
                lines.append(f"- [{title}]({url})")
        
        lines.append("")
    
    return lines

def generate_readme(config, output_path="docs/README.md"):
    """Generate the complete README file"""
    lines = []
    
    # Generate header
    lines.extend(generate_header(config))
    
    # Generate sections
    lines.extend(generate_sections(config))
    
    # Generate quick links
    lines.extend(generate_quick_links(config))
    
    # Write to file
    try:
        os.makedirs(os.path.dirname(output_path), exist_ok=True)
        with open(output_path, 'w', encoding='utf-8') as f:
            f.write('\n'.join(lines))
        print(f"✅ Generated {output_path}")
        return True
    except Exception as e:
        print(f"❌ Error writing {output_path}: {e}")
        return False

def main():
    """Main function"""
    config_path = "docs.yaml"
    output_path = "docs/README.md"
    
    # Load configuration
    config = load_config(config_path)
    if not config:
        sys.exit(1)
    
    # Generate README
    success = generate_readme(config, output_path)
    if not success:
        sys.exit(1)
    
    print("🎉 Documentation README generated successfully!")

if __name__ == "__main__":
    main() 