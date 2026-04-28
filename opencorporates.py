import json
import time
from core.http_client import fetch
from core.utils import save_raw

def run(source_config, query):
    # Build query dynamically from template
    url = source_config["url_template"].replace("{query}", query)

    result = fetch(
        url=url,
        method=source_config.get("method", "GET"),
        rate_limit=source_config.get("rate_limit", 2)
    )

    output = {
        "source_id": source_config["id"],
        "query": query,
        "status": result["status"],
        "timestamp": int(time.time()),
        "response_time": result.get("response_time", 0)
    }

    if result["content"]:
        raw_path = save_raw(source_config["id"], result["content"])
        output["raw_path"] = raw_path

        try:
            data = json.loads(result["content"])
            output["parsed"] = data.get("results", {})
        except (json.JSONDecodeError, TypeError):
            output["parsed"] = None
    else:
        output["error"] = result.get("error")

    return output