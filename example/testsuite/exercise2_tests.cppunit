#include <cppunit/extensions/HelperMacros.h>
extern int fib(int);

class Exercise2Test : public CppUnit::TestFixture {
    CPPUNIT_TEST_SUITE(Exercise2Test);
    CPPUNIT_TEST(FibPos);
    CPPUNIT_TEST(Fib1);
    CPPUNIT_TEST_SUITE_END();

    public:
    void FibPos();
    void Fib1();
};

CPPUNIT_TEST_SUITE_REGISTRATION(Exercise2Test);

void Exercise2Test::FibPos() {
    CPPUNIT_ASSERT_EQUAL(2584, fib(18));
    CPPUNIT_ASSERT_EQUAL(28657, fib(23));
}

void Exercise2Test::Fib1() {
    CPPUNIT_ASSERT_EQUAL(1, fib(1));
}
