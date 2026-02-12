# Formula Engine

### Problem

Complex business systems like Payroll, Pricing, or Risk Scoring involve interdependent calculations (e.g., Net Salary depends on Tax, Tax depends on Bonus).

*   **Rigidity**: Hardcoding formulas in code requires deployment for every change.
*   **Performance**: Naive implementations re-evaluate all formulas (O(N)), wasting resources.
*   **Instability**: Circular dependencies (A -> B -> A) can cause infinite loops and system crashes.

### Solution

A high-performance dynamic formula engine that treats business logic as configuration while maintaining safety and speed.

*   **Dynamic Evaluation**: Formulas are stored as text and executed safely at runtime using `expr`.
*   **Safety (Cycle Detection)**: Uses **Kahn's Algorithm** to build a Dependency Graph and detect infinite loops at startup, preventing runtime crashes.
*   **Performance (O(K))**:
    *   **Initialization**: Pre-calculates an optimized Execution Plan for every formula once.
    *   **Runtime**: Calculates *only* the relevant dependencies for a specific target, reducing complexity from O(N) to O(K).
*   **Traceability**: Returns a full trace of all intermediate values for audit and debugging.
