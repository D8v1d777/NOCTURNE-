import json
import time
from core.http_client import fetch
from core.utils import save_raw

REDDIT_HEADERS = {
    "User-Agent": "nocturne-bot/1.0"
}

def run(source_config, query):
    url = source_config["url_template"].replace("{query}", query)

    result = fetch(
        url=url,
        method=source_config.get("method", "GET"),
        headers=REDDIT_HEADERS,
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

            posts = []
            for item in data.get("data", {}).get("children", []):
                post = item.get("data", {})

                posts.append({
                    "username": post.get("author"),
                    "subreddit": post.get("subreddit"),
                    "title": post.get("title"),
                    "url": f"https://reddit.com{post.get('permalink')}"
                })

            output["parsed"] = posts

        except Exception as e:
            output["parsed"] = None
            output["error"] = str(e)

    else:
        output["error"] = result.get("error")

    return output