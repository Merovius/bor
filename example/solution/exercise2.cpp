// Exercise: Write a recursive version to calculate the n-th
// fibonacci-number

int fib (int n) {
    return (n <= 1) ? 1 : fib(n-1) + fib(n-2);
}
