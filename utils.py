import os
import time

def save_raw(source_id, content):
    timestamp = int(time.time())
    folder = f"./raw/{source_id}"
    os.makedirs(folder, exist_ok=True)

    path = f"{folder}/{timestamp}.html"
    with open(path, "w", encoding="utf-8") as f:
        f.write(content)

    return path