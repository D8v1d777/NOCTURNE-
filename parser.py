import json
import re

class AdaptiveParser:
    """Recursive structural traversal for JSON and HTML data."""

    @staticmethod
    def parse(content, content_type="html"):
        if content_type == "json":
            return AdaptiveParser._parse_json(content)
        return AdaptiveParser._parse_html(content)

    @staticmethod
    def _parse_json(content):
        try:
            data = json.loads(content)
            signals = []

            def traverse(obj, path=""):
                if isinstance(obj, dict):
                    for k, v in obj.items():
                        new_path = f"{path}.{k}" if path else k
                        if isinstance(v, (str, int, float)):
                            signals.append({"key": k, "value": str(v), "context": new_path})
                        else:
                            traverse(v, new_path)
                elif isinstance(obj, list):
                    for i, item in enumerate(obj):
                        traverse(item, f"{path}[{i}]")
            
            traverse(data)
            return signals
        except:
            return []

    @staticmethod
    def _parse_html(content):
        # Strips tags to find text blobs while preserving context for proximity matching
        signals = []
        
        # Extract text nodes with surrounding context
        # This is a lightweight substitute for a full DOM tree
        text_blobs = re.finditer(r'>([^<]{2,})<', content)
        for match in text_blobs:
            text = match.group(1).strip()
            if text:
                # Context is the 50 chars before the tag start
                start = match.start()
                context = content[max(0, start - 50):start]
                signals.append({"key": "text_node", "value": text, "context": context})
        
        # Extract relevant attributes
        links = re.finditer(r'(?:href|src)=["\']([^"\']+)["\']', content)
        for match in links:
            val = match.group(1)
            if val.startswith("http") or "@" in val:
                signals.append({
                    "key": "attribute",
                    "value": val,
                    "context": content[max(0, match.start()-20):match.start()]
                })

        return signals