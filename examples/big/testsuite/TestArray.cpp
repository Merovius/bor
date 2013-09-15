#include <cppunit/extensions/HelperMacros.h>
#include <cstdlib>
#include <ctime>

#include "Bank.hpp"

class AccountsArrayTest : public CppUnit::TestFixture {
    CPPUNIT_TEST_SUITE(AccountsArrayTest);
    CPPUNIT_TEST(TakeEmpty);
    CPPUNIT_TEST(TakeNonExistent);
    CPPUNIT_TEST(PutMoney);
    CPPUNIT_TEST(TakeMoney);
    CPPUNIT_TEST(TransferMoney);
    CPPUNIT_TEST(TransferBetweenBanks);
    CPPUNIT_TEST_SUITE_END();

    private:
    Bank *bank1;
    Bank *bank2;
    AccountNumber beate;
    AccountNumber harald;
    AccountNumber charlie;

    public:
    void setUp();
    void TakeEmpty();
    void TakeNonExistent();
    void PutMoney();
    void TakeMoney();
    void TransferMoney();
    void TransferBetweenBanks();
    void tearDown();
};

CPPUNIT_TEST_SUITE_REGISTRATION(AccountsArrayTest);

void AccountsArrayTest::setUp() {
    bank1 = new Bank("TestBank", "array");
    beate = bank1->createAccount("Beate");
    harald = bank1->createAccount("Harald");
    bank2 = new Bank("TestBank2", "array");
    charlie = bank2->createAccount("Charlie");
    srand(time(NULL));
}

void AccountsArrayTest::TakeEmpty() {
    // Random amount between 1 and 100
    int n = rand() % 100 + 1;

    CPPUNIT_ASSERT_EQUAL(bank1->takeMoney(beate, n), 0);
}

void AccountsArrayTest::TakeNonExistent() {
    int acc;

    // Generate a random account not taken yet
    while (1) {
        acc = rand() % 10000 + 1;
        if (acc != beate && acc != harald) {
            break;
        }
    }

    // Take a random amount between 1 and 100 from the account
    int n = rand() % 100 + 1;
    CPPUNIT_ASSERT_EQUAL(bank1->takeMoney(acc, n), 0);
}

void AccountsArrayTest::PutMoney() {
    // Put a random amount between 1 and 100 into the account
    int n = rand() % 100 + 1;
    CPPUNIT_ASSERT(bank1->putMoney(beate, n));

    int m = rand() % 100 + 1;
    CPPUNIT_ASSERT(bank1->putMoney(beate, m));

    CPPUNIT_ASSERT_EQUAL(bank1->balance(beate), n+m);
}

void AccountsArrayTest::TakeMoney() {
    // We have to put something in, before taking it out
    int n = rand() % 100 + 101;
    CPPUNIT_ASSERT(bank1->putMoney(harald, n));

    // Take a random amount between 1 and 100 out
    int m = rand() % 100 + 1;
    CPPUNIT_ASSERT(bank1->takeMoney(harald, m));

    CPPUNIT_ASSERT_EQUAL(bank1->balance(harald), n - m);
}

void AccountsArrayTest::TransferMoney() {
    int balanceBeate = bank1->balance(beate);
    int balanceHarald = bank1->balance(harald);

    // We add some money to beates account and transfer (some of) it to haralds
    int n = rand() % 100 + 101;
    CPPUNIT_ASSERT(bank1->putMoney(beate, n));

    int m = rand() % 100 + 1;
    CPPUNIT_ASSERT(bank1->transferMoney(beate, m, bank1, harald));

    CPPUNIT_ASSERT_EQUAL(bank1->balance(beate), balanceBeate + n - m);
    CPPUNIT_ASSERT_EQUAL(bank1->balance(harald), balanceHarald + m);
}

void AccountsArrayTest::TransferBetweenBanks() {
    int balanceBeate = bank1->balance(beate);
    int balanceCharlie = bank2->balance(charlie);

    // We add some money to charlies account and transfer (some of) it to beates
    int n = rand() % 100 + 101;
    CPPUNIT_ASSERT(bank2->putMoney(charlie, n));

    int m = rand() % 100 + 1;
    CPPUNIT_ASSERT(bank2->transferMoney(charlie, m, bank1, beate));

    CPPUNIT_ASSERT_EQUAL(bank2->balance(charlie), balanceCharlie + n - m);
    CPPUNIT_ASSERT_EQUAL(bank1->balance(beate), balanceBeate + m);
}

void AccountsArrayTest::tearDown() {
    delete bank1;
    delete bank2;
}
