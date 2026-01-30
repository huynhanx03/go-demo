# Promotion Code Validation Service

High-performance Go service for validating promotion codes using **Bloom Filters** and **Hexagonal Architecture**.

## 1. Overview & Approach

### The Problem
We need to validate if a user input code exists in **BOTH** `campaign_codes.txt` and `membership_codes.txt`.
Constraint: Files are large (millions of records), requiring an efficient solution for memory and speed.

### The Solution: **Intersection Preprocessing**
Instead of checking two files at runtime or loading everything into a map (high RAM usage), we use a **Bloom Filter**:
1.  **Startup Phase**:
    *   Load `campaign_codes.txt` into a temporary Bloom Filter.
    *   Stream `membership_codes.txt`, checking each code against the temporary filter.
    *   If a match is found (Code ∈ Campaign AND Code ∈ Member), add it to the **Final Bloom Filter**.
2.  **Runtime Phase**:
    *   `IsEligible(code)` simply checks the Final Bloom Filter.
    *   Result: **O(1)** complexity, extremely fast.

## 2. Data Structures

*   **Bloom Filter**:
    *   A probabilistic data structure used to test whether an element is a member of a set.
    *   **Why?** It consumes significantly less memory than a Hash Map.
    *   **Optimization**: We use a bitset array and multiple hash functions to minimize collisions.

## 3. Performance Considerations

*   **Memory Efficiency**: By using Bloom Filter, we only use ~30MB of RAM to represent millions of active codes, compared to GBs if using `map[string]bool`.
*   **Startup vs Runtime**:
    *   We trade **Startup Time** (~5s to scan files) for **Zero-Latency Runtime**.
    *   Once loaded, validation is instantaneous.
*   **Concurrency**:
    *   `gen-data` uses a **Worker Pool** pattern to generate 10M codes in parallel, utilizing all CPU cores.

## 4. Assumptions

1.  **Data Integrity**: We assume the input files are text files with one code per line.
2.  **False Positives**: Bloom Filters have a small theoretical chance of False Positives (saying a code is valid when it's not), but False Negatives are impossible. We configured the filter size to keep this rate negligible (~0.0001%).
3.  **Code Format**: Codes are strings (1-5 characters, a-z).

---

## 5. How to Run

### Prerequisites
*   Go 1.22+
*   Make

### Step 1: Generate Data
Generate 10 million random codes (Campaign, Member, and Test Cases).
```bash
make gen-data
```

### Step 2: Run Analyzer (Batch Mode)
Verify the system accuracy and measure performance statistics.
```bash
make analyze
```
*Output includes: Load Time, Process Time, and Accuracy (100%).*

### Step 3: Run Checker (Interactive Mode)
Manually test specific codes.
```bash
make check
```
*   Type valid code (e.g., from `data/test_cases.txt` marked true).
*   Type invalid code.
*   Output is colored (Green/Red) for easy visibility.
