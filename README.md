# Task: Design an Efficient In-Memory Search Engine for Text Documents

## Problem Statement:

You are tasked with building a simple in-memory search engine that allows users to upload text documents and perform efficient searches on them. The search engine should allow users to find documents that contain specific words or phrases, with an emphasis on efficiency and scalability for large datasets.

## Requirements:

### Document Storage:

Implement an API that accepts and stores text documents in memory.
Each document should be assigned a unique document ID.
Search Functionality:

Implement an efficient search algorithm that allows users to search for specific words or phrases within the documents.
The search should be case-insensitive.
The search should support both exact match (whole words) and partial match (substrings).
Indexing:

To enhance search performance, implement a basic indexing system that preprocesses the documents and allows for fast lookups.
The indexing system should update dynamically as new documents are added or removed.

### Endpoints:

`POST /documents`: To upload a text document.
`GET /documents/search?query=your_query`: To search for a word or phrase across all documents. Return a list of document IDs where the search term appears.
`DELETE /documents/:id`: To remove a document by its ID from the system.

### Constraints:

Design for scalability: Assume there may be thousands of documents in the system, and searches need to be fast.
Optimize memory usage, ensuring the system can handle a large number of documents without consuming excessive memory.
Advanced Features (Optional if time permits):

Add support for Boolean operators in search queries (e.g., "AND", "OR", "NOT").
Implement a way to rank results by relevance based on how frequently the search term appears in each document.

## Reflection:

After completing the task, ask him to document his design choices for indexing and optimizing search speed, and discuss potential improvements for larger-scale systems.
Technical Focus:

This task challenges him to think critically about data structure design, efficient searching algorithms (such as using hash maps or tries), and memory optimization. It's complex enough to engage his problem-solving skills but should be achievable in around twenty minutes with his experience.