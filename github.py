import json
import time
from core.http_client import fetch
from core.utils import save_raw

GITHUB_HEADERS = {
    "Accept": "application/vnd.github+json"
}

def run(source_config, query):
    url = source_config["url_template"].replace("{query}", query)

    result = fetch(
        url=url,
        method=source_config.get("method", "GET"),
        headers=GITHUB_HEADERS,
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

            users = []
            for item in data.get("items", []):
                users.append({
                    "username": item.get("login"),
                    "profile_url": item.get("html_url"),
                    "avatar": item.get("avatar_url"),
                    "score": item.get("score")
                })

            output["parsed"] = users

        except Exception as e:
            output["parsed"] = None
            output["parse_error"] = str(e)

    else:
        output["error"] = result.get("error")

    return output