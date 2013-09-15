#ifndef _BANK_HPP
#define _BANK_HPP

using namespace std;

#include "Account.hpp"
#include "Accounts.hpp"
#include "AccountsArray.hpp"
#include "AccountsList.hpp"

class Bank {
	string name;
	string mode; // either array or list
	Accounts* accounts;
public:
	Bank(string n, string m);
	AccountNumber createAccount(string name);
	Euro takeMoney(AccountNumber n, Euro b);
	bool putMoney(AccountNumber n, Euro b);
	bool transferMoney(AccountNumber n1, Euro e, Bank* b, AccountNumber n2);
	Euro balance(AccountNumber n);
};

#endif /* _BANK_HPP */
