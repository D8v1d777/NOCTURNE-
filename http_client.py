import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry
import random
import time

# Rotated User-Agents to mimic modern browsers and avoid basic fingerprinting
USER_AGENTS = [
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
    "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36"
]

# Initialize a shared session with retry logic for efficiency and resilience
_session = requests.Session()
retry_strategy = Retry(
    total=3,
    backoff_factor=1,
    status_forcelist=[429, 500, 502, 503, 504],
)
adapter = HTTPAdapter(max_retries=retry_strategy)
_session.mount("http://", adapter)
_session.mount("https://", adapter)

def fetch(url, method="GET", headers=None, rate_limit=2, timeout=15):
    """
    Performs a resilient HTTP request with automated retries and rotated identity.
    """
    if headers is None:
        headers = {}

    # Apply rotation and jittered rate limiting
    headers["User-Agent"] = random.choice(USER_AGENTS)
    time.sleep(random.uniform(rate_limit, rate_limit + 2))

    try:
        response = _session.request(
            method=method, 
            url=url, 
            headers=headers, 
            timeout=timeout
        )
        return {
            "status": response.status_code,
            "content": response.text,
            "response_time": response.elapsed.total_seconds()
        }
    except Exception as e:
        return {
            "status": "error",
            "error": str(e),
            "content": None
        }