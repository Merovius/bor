// Exercise: Write a recursive function, calculating the binomial
// coefficient n choose k

int binom(int n, int k) {
    return (n < k) ? 0 :
        (n == k) ? 1 : (binom(n-1, k) * n / (n-k));
}
