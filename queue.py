import redis
import json

class NocturneQueue:
    """Redis-backed queue for distributing scanning tasks and aggregating results."""
    
    def __init__(self, host='localhost', port=6379, db=0):
        self.client = redis.Redis(host=host, port=port, db=db, decode_responses=True)
        self.job_key = "nocturne:jobs"
        self.result_key = "nocturne:results"

    def push_job(self, task, custom_key=None):
        """Add a scan task to the queue."""
        key = custom_key or self.job_key
        self.client.lpush(key, json.dumps(task))

    def pop_job(self, timeout=0, custom_key=None):
        """Block until a job is available and return it."""
        key = custom_key or self.job_key
        job = self.client.brpop(key, timeout=timeout)
        if job:
            # brpop returns a tuple (key, value)
            return json.loads(job[1])
        return None

    def push_result(self, result):
        """Add a completed scan result to the aggregator queue."""
        self.client.lpush(self.result_key, json.dumps(result))

    def pop_result(self, timeout=0):
        """Block until a result is available and return it."""
        res = self.client.brpop(self.result_key, timeout=timeout)
        if res:
            return json.loads(res[1])
        return None