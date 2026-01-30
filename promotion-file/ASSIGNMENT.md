# Promotion Code Validation

## Business Context

Our platform manages **promotion codes** across two independent data sources:

- **Campaign System**
- **Membership System**

Your task is to determine whether a given promotion code is **eligible** for use.

---

## Eligible Promotion Code

A promotion code is **eligible** if it exists in **both** data sources.

---

## Input Data

### 1. Data Sources

Two large text files (input files may not fit entirely into memory):

- `campaign_codes.txt`
- `membership_codes.txt`

Each file:

- Contains **millions of promotion codes** (one per line)
- Codes are **unique within each file**
- Code format: **max 5 characters**, lowercase letters (`a-z`)

### 2. Code to Validate

- A string `code` entered by the user
- Constraints: `1 ≤ length ≤ 5`, characters `a-z`

## Output

- `true` if the code exists in **both** data sources
- `false` otherwise

## Example

**campaign_codes.txt**
```
abc
xyz
promo
sale
```

**membership_codes.txt**
```
abc
gold
promo
```

**Input:** `code = "promo"`

**Output:** `true`