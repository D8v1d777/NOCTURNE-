import hashlib

def generate_id(value):
    return hashlib.sha256(value.encode()).hexdigest()[:16]


# =========================
# GITHUB NORMALIZER
# =========================

def normalize_github(data):
    identities = []

    for user in data.get("parsed", []):
        uid = generate_id(user.get("username", ""))

        identities.append({
            "id": uid,
            "username": user.get("username"),
            "name": None,
            "platform": "github",
            "source": "github",
            "bio": None,
            "location": None,
            "links": [user.get("profile_url")],
            "confidence": 0.7,
            "raw_source": data.get("raw_path")
        })

    return identities


# =========================
# OPENCORPORATES NORMALIZER
# =========================

def normalize_opencorporates(data):
    identities = []

    results = data.get("parsed", {}).get("companies", [])

    for item in results:
        company = item.get("company", {})
        name = company.get("name", "")

        uid = generate_id(name)

        identities.append({
            "id": uid,
            "username": None,
            "name": name,
            "platform": "corporate",
            "source": "opencorporates",
            "bio": None,
            "location": company.get("registered_address"),
            "links": [company.get("opencorporates_url")],
            "confidence": 0.6,
            "raw_source": data.get("raw_path")
        })

    return identities


# =========================
# REDDIT NORMALIZER
# =========================

def normalize_reddit(data):
    identities = []

    for post in data.get("parsed", []):
        username = post.get("username")
        if not username:
            continue

        uid = generate_id(username)

        identities.append({
            "id": uid,
            "username": username,
            "name": None,
            "platform": "reddit",
            "source": "reddit",
            "bio": post.get("title"),
            "location": None,
            "links": [post.get("url")],
            "confidence": 0.65,
            "raw_source": data.get("raw_path")
        })

    return identities


# =========================
# MASTER NORMALIZER
# =========================

def normalize(source_id, data):
    if source_id == "github":
        return normalize_github(data)

    elif source_id == "opencorporates":
        return normalize_opencorporates(data)

    elif source_id == "reddit":
        return normalize_reddit(data)

    else:
        return []