from pymongo import MongoClient

client = MongoClient("mongodb://localhost:27017/")
db = client["testdb"]
employees = db["employees"]
channels = db["channels"]

# ---- Section Level ----
section_pipeline = [
    {"$match": {"division_id": {"$in": ["A", "B"]}}},
    {"$group": {
        "_id": {
            "division_id": "$division_id",
            "department_id": "$department_id",
            "section_id": "$section_id"
        },
        "members": {"$addToSet": "$account_id"}
    }},
    {"$project": {
        "_id": 0,
        "type": {"$literal": "section"},
        "division_id": "$_id.division_id",
        "department_id": "$_id.department_id",
        "section_id": "$_id.section_id",
        "members": 1
    }}
]

for doc in employees.aggregate(section_pipeline):
    query = {
        "type": "section",
        "division_id": doc["division_id"],
        "department_id": doc["department_id"],
        "section_id": doc["section_id"]
    }
    channels.update_one(query, {"$set": doc}, upsert=True)

# ---- Department Level ----
department_pipeline = [
    {"$match": {"division_id": {"$in": ["A", "B"]}}},
    {"$group": {
        "_id": {
            "division_id": "$division_id",
            "department_id": "$department_id"
        },
        "members": {"$addToSet": "$account_id"}
    }},
    {"$project": {
        "_id": 0,
        "type": {"$literal": "department"},
        "division_id": "$_id.division_id",
        "department_id": "$_id.department_id",
        "members": 1
    }}
]

for doc in employees.aggregate(department_pipeline):
    query = {
        "type": "department",
        "division_id": doc["division_id"],
        "department_id": doc["department_id"]
    }
    channels.update_one(query, {"$set": doc}, upsert=True)
