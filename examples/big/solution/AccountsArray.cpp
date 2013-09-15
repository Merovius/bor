using namespace std;

#include "AccountsArray.hpp"

AccountsArray::AccountsArray() {
	freeaccount = 0;
}

AccountsArray::~AccountsArray() {
	delete accountset;
}

/**
 * Create an account and decrease the amount of available accounts
 * Return the account number of created account
 */
AccountNumber AccountsArray::createAccount(string name) {
	accountset[freeaccount] = Account(0, name, freeaccount);
	freeaccount++;

	return freeaccount - 1;
}

/**
 * Search for an account and return either true for found or false
 */
ResultFindAccount AccountsArray::findAccount(AccountNumber n) {
	ResultFindAccount f;
	if (n < freeaccount) {
		f.r = true;
		f.a = &accountset[n];
	} else {
		f.r = false;
	}
	return f;
}
