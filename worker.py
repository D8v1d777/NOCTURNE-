import time
import json
import requests
import random
import sys
import os
from queue import NocturneQueue
from source_profile import SourceProfileManager
from core.classifier import SourceClassifier
from core.parser import AdaptiveParser
from core.detector import FieldDetector
from core.entity_builder import EntityBuilder

class Worker:
    """
    Stateless geo-distributed worker with adaptive request handling.
    Capabilities: 'html' (requests), 'js' (playwright), 'api'
    """
    
    def __init__(self, worker_id, region="US", capability="html"):
        self.id = worker_id
        self.region = region
        self.capability = capability
        self.queue = NocturneQueue()
        self.profiles = SourceProfileManager()
        self.user_agents = [
            "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
            "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36"
        ]
        
        # Initialize Playwright only if needed
        self.browser = None
        if capability == "js":
            self._init_js_engine()

    def _init_js_engine(self):
        print(f"[*] Initializing Playwright engine for Worker {self.id}")
        # Placeholder for Playwright setup logic

    def start(self):
        routing_key = f"nocturne:jobs:{self.region}:{self.capability}"
        print(f"[*] Worker [{self.id}] ({self.region}/{self.capability}) listening on {routing_key}")
        while True:
            # Workers listen to their specific geo/cap queue
            job = self.queue.pop_job(custom_key=routing_key)
            if job:
                self.process_job(job)

    def process_job(self, job):
        source_id = job['source_id']
        target = job['target']
        url = job['url_template'].format(query=target)
        
        # 4. ADAPTIVE REQUEST ENGINE: Dynamic Delay + Jitter
        profile = self.profiles.get_profile(source_id)
        base_delay = profile.get("recommended_delay", 2.0)
        jitter = random.uniform(0.8, 1.2) # 20% variance
        
        print(f"[{self.id}] Adaptive Wait: {base_delay * jitter:.2f}s for {source_id}")
        time.sleep(base_delay * jitter)

        start_time = time.time()
        try:
            headers = {"User-Agent": random.choice(self.user_agents)}
            
            if self.capability == "js":
                # Execute via Playwright (Simulated here)
                response = requests.get(url, headers=headers, timeout=20)
            else:
                response = requests.get(url, headers=headers, timeout=15)
            
            # 2. Adaptive Pipeline
            status_code = response.status_code
            raw_content = response.text
            execution_time = time.time() - start_time

            # Update Source Intelligence
            self.profiles.update_stats(source_id, status_code, execution_time)

            if status_code >= 400:
                self.handle_failure(job, f"HTTP {status_code}")
                return

            classification = SourceClassifier.classify(source_id, raw_content)
            signals = AdaptiveParser.parse(raw_content, classification["content_type"])
            detected_fields = FieldDetector.detect(signals)
            identities = EntityBuilder.build(source_id, detected_fields)

            # Push to results
            self.queue.push_result({
                "source_id": source_id,
                "target": target,
                "identities": identities,
                "execution_time": round(execution_time, 3),
                "status": "success"
            })
            print(f"[+] [{self.id}] Success: {source_id} ({self.region})")

        except Exception as e:
            self.handle_failure(job, str(e))

    def handle_failure(self, job, error_msg):
        """4. Retry System with Exponential Backoff."""
        retries = job.get("retries", 0)
        if retries < 3:
            job["retries"] = retries + 1
            # Exponential backoff: 2, 4, 8 seconds
            backoff = (2 ** job["retries"]) + random.uniform(0, 1)
            print(f"[!] [{self.id}] Error on {job['source_id']}: {error_msg}. Retrying in {backoff:.2f}s...")
            time.sleep(backoff)
            self.queue.push_job(job, custom_key=f"nocturne:jobs:{self.region}:{self.capability}")
        else:
            print(f"[X] [{self.id}] Failed {job['source_id']} after max retries.")
            self.queue.push_result({
                "source_id": job["source_id"],
                "target": job["target"],
                "status": "failed",
                "error": error_msg
            })

if __name__ == "__main__":
    worker_id = os.getenv("WORKER_ID", f"worker-{random.randint(1000, 9999)}")
    region = os.getenv("REGION", "US")
    capability = os.getenv("CAPABILITY", "html")

    sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    
    w = Worker(worker_id, region, capability)
    w.start()