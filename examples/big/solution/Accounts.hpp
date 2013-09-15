#ifndef _ACCOUNTS_HPP
#define _ACCOUNTS_HPP

#include "Account.hpp"

using namespace std;

struct ResultFindAccount {
    bool r;
    Account* a;
};

class Accounts {
public:
	virtual AccountNumber createAccount(string name) = 0;
	virtual ResultFindAccount findAccount(AccountNumber accountNumber) = 0;
};

#endif /* _ACCOUNTS_HPP */
