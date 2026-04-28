import time
import sys

if __name__ == "__main__":
    region = sys.argv[1] if len(sys.argv) > 1 else "local"
    mode = sys.argv[2] if len(sys.argv) > 2 else "html"

    print(f"[WORKER] Started ({region}/{mode})")

    while True:
        print(f"[WORKER] Waiting for tasks...")
        time.sleep(5)