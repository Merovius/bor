#include <cppunit/extensions/HelperMacros.h>
extern int binom(int, int);

class Exercise1Test : public CppUnit::TestFixture {
    CPPUNIT_TEST_SUITE(Exercise1Test);
    CPPUNIT_TEST(PosChoosePos);
    CPPUNIT_TEST(PosChooseO);
    CPPUNIT_TEST(PosChooseSame);
    CPPUNIT_TEST(PosChooseGreater);
    CPPUNIT_TEST_SUITE_END();

    public:
    void PosChoosePos();
    void PosChooseO();
    void PosChooseSame();
    void PosChooseGreater();
};

CPPUNIT_TEST_SUITE_REGISTRATION(Exercise1Test);

void Exercise1Test::PosChoosePos() {
    CPPUNIT_ASSERT_EQUAL(binom(20, 5), 15504);
    CPPUNIT_ASSERT_EQUAL(binom(13, 11), 78);
}

void Exercise1Test::PosChooseO() {
    CPPUNIT_ASSERT_EQUAL(binom(20, 0), 1);
    CPPUNIT_ASSERT_EQUAL(binom(13, 0), 1);
}

void Exercise1Test::PosChooseSame() {
    CPPUNIT_ASSERT_EQUAL(binom(20, 20), 1);
    CPPUNIT_ASSERT_EQUAL(binom(13, 13), 1);
}

void Exercise1Test::PosChooseGreater() {
    CPPUNIT_ASSERT_EQUAL(binom(20, 21), 0);
    CPPUNIT_ASSERT_EQUAL(binom(13, 18), 0);
}
