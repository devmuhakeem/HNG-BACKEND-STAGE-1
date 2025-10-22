# HNG-Stage-1: String Analyzer API 🔍

## Overview
This backend service is developed in Go, leveraging its standard `net/http` package to provide a robust RESTful API. It enables efficient storage, detailed analysis, and flexible retrieval of string data, including unique features like natural language query processing for filtering.

## Features
- **Comprehensive String Analysis**: Automatically computes string length, identifies palindromes, counts unique characters, determines word count, and generates SHA256 hashes for each stored string.
- **Ephemeral Data Store**: Utilizes a concurrent in-memory map for quick storage and retrieval of analyzed string objects, suitable for high-performance temporary data handling.
- **Advanced Filtering Capabilities**: Supports detailed filtering of stored strings based on various properties such as palindrome status, length ranges, specific word counts, and the presence of particular characters via query parameters.
- **Natural Language Query Processing**: Allows users to interact with the API using descriptive, human-readable sentences to define complex filtering criteria, enhancing user experience.
- **RESTful API Design**: Implements standard HTTP methods (POST, GET, DELETE) for intuitive and predictable interaction with string resources.

## Getting Started
To get the String Analyzer API up and running on your local machine, follow these steps:

### Installation
- **Clone the Repository:**
  Begin by cloning the project repository to your local machine.
  ```bash
  git clone https://github.com/samueltuoyo15/HNG-Stage-1.git
  cd HNG-Stage-1
  ```
- **Initialize Go Modules:**
  Ensure all necessary Go modules are downloaded and synchronized.
  ```bash
  go mod tidy
  ```
- **Run the Application:**
  Start the API server.
  ```bash
  go run main.go
  ```
  The API server will become accessible at `http://localhost:8080`.

### Environment Variables
This project does not require any specific environment variables for its default operation. The server binds to port `8080` by default.

## API Documentation
### Base URL
`http://localhost:8080`

### Endpoints

#### `POST /strings`
**Description**: Stores a new string value and returns its computed analysis. If the string (identified by its SHA256 hash) already exists, a conflict error is returned.

**Request**:
```json
{
  "value": "your string here"
}
```
**Required Fields**:
- `value` (string): The string to be analyzed and stored.

**Response**:
```json
{
  "id": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "value": "your string here",
  "properties": {
    "length": 16,
    "is_palindrome": false,
    "unique_characters": 9,
    "word_count": 3,
    "sha256_hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    "character_frequency_map": {
      " ": 2,
      "e": 2,
      "g": 1,
      "h": 2,
      "i": 1,
      "n": 1,
      "o": 1,
      "r": 2,
      "s": 1,
      "t": 1,
      "u": 1,
      "y": 1
    }
  },
  "created_at": "2023-10-27T10:00:00Z"
}
```
**Errors**:
- `400 Bad Request`: Invalid JSON body or missing `value` field.
- `422 Unprocessable Entity`: The `value` field is not a string.
- `409 Conflict`: The string already exists in the system.

#### `GET /strings`
**Description**: Retrieves all stored strings, with optional filtering capabilities based on various properties via query parameters.

**Request**:
Query Parameters:
- `is_palindrome` (boolean, optional): Filters strings by their palindrome status (`true` or `false`).
- `min_length` (integer, optional): Filters for strings with a length greater than or equal to this value.
- `max_length` (integer, optional): Filters for strings with a length less than or equal to this value.
- `word_count` (integer, optional): Filters for strings that have an exact number of words.
- `contains_character` (string, optional): Filters for strings that contain this single specified character.

**Response**:
```json
{
  "data": [
    {
      "id": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
      "value": "your string here",
      "properties": {
        "length": 16,
        "is_palindrome": false,
        "unique_characters": 9,
        "word_count": 3,
        "sha256_hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
        "character_frequency_map": {
          " ": 2,
          "e": 2,
          "g": 1,
          "h": 2,
          "i": 1,
          "n": 1,
          "o": 1,
          "r": 2,
          "s": 1,
          "t": 1,
          "u": 1,
          "y": 1
        }
      },
      "created_at": "2023-10-27T10:00:00Z"
    }
  ],
  "count": 1,
  "filters_applied": {
    "is_palindrome": false,
    "min_length": 10
  }
}
```
**Errors**:
- `400 Bad Request`: An invalid value was provided for any query parameter (e.g., non-boolean for `is_palindrome`, non-integer for lengths, or more than one character for `contains_character`).

#### `GET /strings/{value}`
**Description**: Retrieves the details of a specific string by providing its original value. The `{value}` in the path must be URL-encoded.

**Request**:
Path Parameter:
- `{value}` (string): The URL-encoded original string to retrieve.

**Response**:
```json
{
  "id": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "value": "your string here",
  "properties": {
    "length": 16,
    "is_palindrome": false,
    "unique_characters": 9,
    "word_count": 3,
    "sha256_hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    "character_frequency_map": {
      " ": 2,
      "e": 2,
      "g": 1,
      "h": 2,
      "i": 1,
      "n": 1,
      "o": 1,
      "r": 2,
      "s": 1,
      "t": 1,
      "u": 1,
      "y": 1
    }
  },
  "created_at": "2023-10-27T10:00:00Z"
}
```
**Errors**:
- `400 Bad Request`: Invalid URL-encoded string or a missing string value in the path.
- `404 Not Found`: The string does not exist in the system.

#### `GET /strings/filter-by-natural-language`
**Description**: Filters stored strings using a natural language query provided as a parameter. The API attempts to parse the query into filter criteria.

**Request**:
Query Parameter:
- `query` (string): A natural language sentence describing the desired string properties (e.g., "strings longer than 5 characters and containing the letter a").

**Response**:
```json
{
  "data": [
    {
      "id": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
      "value": "your string here",
      "properties": {
        "length": 16,
        "is_palindrome": false,
        "unique_characters": 9,
        "word_count": 3,
        "sha256_hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
        "character_frequency_map": {
          " ": 2,
          "e": 2,
          "g": 1,
          "h": 2,
          "i": 1,
          "n": 1,
          "o": 1,
          "r": 2,
          "s": 1,
          "t": 1,
          "u": 1,
          "y": 1
        }
      },
      "created_at": "2023-10-27T10:00:00Z"
    }
  ],
  "count": 1,
  "interpreted_query": {
    "original": "find strings longer than 5 characters containing the letter r",
    "parsed_filters": {
      "contains_character": "r",
      "min_length": 6
    }
  }
}
```
**Errors**:
- `400 Bad Request`: Missing `query` parameter or the natural language query cannot be parsed into valid filters.
- `422 Unprocessable Entity`: The natural language query contains conflicting filters (e.g., specifying a minimum length greater than a maximum length).

#### `DELETE /strings/{value}`
**Description**: Deletes a specific string from the in-memory store by its original value. The `{value}` in the path must be URL-encoded.

**Request**:
Path Parameter:
- `{value}` (string): The URL-encoded original string to be deleted.

**Response**:
`204 No Content` (No response body for a successful deletion)

**Errors**:
- `400 Bad Request`: Invalid URL-encoded string or a missing string value in the path.
- `404 Not Found`: The string does not exist in the system.

---

## Usage
Once the API server is running, you can interact with it using `curl` or any other HTTP client. Here are some common use cases:

- **Add a new string:**
  ```bash
  curl -X POST -H "Content-Type: application/json" -d '{"value": "Hello world"}' http://localhost:8080/strings
  ```

- **Retrieve all strings:**
  ```bash
  curl http://localhost:8080/strings
  ```

- **Retrieve all palindromic strings longer than 5 characters:**
  ```bash
  curl "http://localhost:8080/strings?is_palindrome=true&min_length=6"
  ```

- **Retrieve a specific string by its value (URL-encoded):**
  ```bash
  curl "http://localhost:8080/strings/Hello%20world"
  ```

- **Filter strings using a natural language query:**
  ```bash
  curl "http://localhost:8080/strings/filter-by-natural-language?query=strings%20longer%20than%205%20characters%20containing%20the%20letter%20a"
  ```

- **Delete a specific string by its value (URL-encoded):**
  ```bash
  curl -X DELETE "http://localhost:8080/strings/Hello%20world"
  ```

## Technologies Used
| Technology      | Description                                                 |
| :-------------- | :---------------------------------------------------------- |
| Go 1.24.3       | The primary programming language used for backend development. |
| `net/http`      | Go's standard library for building high-performance HTTP servers. |
| `encoding/json` | Utilized for efficient JSON serialization and deserialization. |
| `crypto/sha256` | Employed for generating secure cryptographic hashes of string values. |
| `regexp`        | Used for regular expression-based parsing, particularly in natural language processing. |
| `sync`          | Provides primitives for safe and efficient concurrent access to the in-memory data store. |

## License
License information for this project is not explicitly defined in the repository.

## Author Info
👤 **Samuel Tuoyo**
*   Twitter: [@TuoyoS26091](https://x.com/TuoyoS26091)

## Badges
[![Go Reference](https://pkg.go.dev/badge/github.com/samueltuoyo15/HNG-Stage-1.svg)](https://pkg.go.dev/github.com/samueltuoyo15/HNG-Stage-1)
[![Go Report Card](https://goreportcard.com/badge/github.com/samueltuoyo15/HNG-Stage-1)](https://goreportcard.com/report/github.com/samueltuoyo15/HNG-Stage-1)

[![Readme was generated by Dokugen](https://img.shields.io/badge/Readme%20was%20generated%20by-Dokugen-brightgreen)](https://www.npmjs.com/package/dokugen)