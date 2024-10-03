# Task: Build an In-Memory Search Engine with Synonym Support, Ranking, and Proximity Search

## Problem Statement:

You are tasked with building an advanced in-memory search engine that allows users to upload text documents and perform complex searches, including support for synonyms, ranking results by relevance, and proximity-based searches.

## Requirements:

### Document Storage:

Implement an API that accepts and stores text documents in memory.
Each document should be assigned a unique document ID.
Support documents of varying lengths, including very large documents (e.g., thousands of words).

### Search Functionality:

Implement a full-text search system that allows users to search for specific words or phrases within the documents.
The search must be case-insensitive.
The search should support both exact match (whole words) and partial match (substrings).
Additionally, implement proximity search, where users can find documents containing two or more terms within a specified distance (e.g., within 5 words of each other).

### Synonym Support:

Enhance the search engine with synonym support. For example, if a user searches for "quick," the engine should also return documents containing synonyms like "fast."
Implement a basic in-memory dictionary or use a pre-defined set of synonyms for testing (you don’t need to create a huge thesaurus, just enough to showcase functionality).
Indexing and Optimization:

Implement a dynamic indexing system that preprocesses and indexes the documents for fast search operations.
Optimize for efficient memory usage and quick search results. The system should be able to handle a large number of documents efficiently, with minimal latency in searches.
Use an appropriate data structure (like a trie, inverted index, or hash map) to store and query documents efficiently.
Ranking System:

Implement a ranking algorithm to prioritize documents based on relevance. Consider factors like:
Frequency of the search term within the document.
Proximity of search terms to one another (if proximity search is used).
Length of the document (shorter, more focused documents may be ranked higher).
Return search results sorted by relevance score.

### Endpoints:

- `POST /doc`: Upload a new text document.
- `GET /doc/search?q=your_query[&proximity=x]`: Search for documents containing the query. Optionally, if a proximity parameter is provided, return documents where the terms appear within the specified proximity. **See `https://www.sqlite.org/fts5.html` for query syntax.**
- `GET /doc/search?q=your_query[&synonyms=true]`: Search for documents including both the query and its synonyms. **Not implemented.**
- `DELETE /doc/:id`: Remove a document by its ID from the system.
Error Handling and Edge Cases:

Handle cases where no documents match the search.
Handle very large documents and ensure the system does not run out of memory.
Ensure that the indexing system is updated correctly when documents are added, updated, or deleted.

### Advanced Features (Optional):

- Add support for wildcard searches (e.g., searching for “dev*” would return documents containing “developer,” “development,” etc.).
- Implement fuzzy search to account for minor misspellings or typos in the query (e.g., searching for "color" should return results with "colour").
- Add a time complexity analysis for various operations like search, add, and delete, to reflect on efficiency.

### Recording and Reflection:

After completing the task, ask him to document the data structures used, how he handled synonym support, and the logic behind the ranking system.
Have him reflect on potential ways to scale the system if more documents or more complex queries were added.
