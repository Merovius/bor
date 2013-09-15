#ifndef _ACCOUNTS_ARRAY_HPP
#define _ACCOUNTS_ARRAY_HPP

using namespace std;

#include "Accounts.hpp"

class AccountsArray : public Accounts {
private:
	Account accountset[1000];
	AccountNumber freeaccount;
public:
	AccountsArray();
	~AccountsArray();
	AccountNumber createAccount(string);
	ResultFindAccount findAccount(AccountNumber);
};

#endif /* _ACCOUNTS_ARRAY_HPP */
