import json
import os
import sys
from adapters import opencorporates, github, reddit
from ingestion.core.normalizer import normalize

def load_sources():
    # Note: This simplified path assumes sources.json is in the same directory as main.py
    # For more robust path handling, consider using os.path.join(os.path.dirname(__file__), "sources.json")
    # or a similar approach to locate the file relative to the script's execution.
    with open("sources.json") as f:
        return json.load(f)["sources"]

def main():
    sources = load_sources()
    query = "johnsmith"

    all_identities = []

    for source in sources:
        if source["id"] == "opencorporates":
            result = opencorporates.run(source, query)
        elif source["id"] == "github":
            result = github.run(source, query)
        elif source["id"] == "reddit":
            result = reddit.run(source, query)
        else:
            continue

        # Normalize and output structured identities
        identities = normalize(source["id"], result)
        all_identities.extend(identities)

    print(json.dumps(all_identities, indent=2))

if __name__ == "__main__":
    main()