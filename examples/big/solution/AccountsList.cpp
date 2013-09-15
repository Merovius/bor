#ifndef _ACCOUNTS_LIST_HPP
#define _ACCOUNTS_LIST_HPP

using namespace std;

#include "AccountsList.hpp"

AccountsList::AccountsList() {
	accountset = NULL;
	freeaccount = 0;
}

AccountsList::~AccountsList() {
	delete accountset;
}

AccountNumber AccountsList::createAccount(string name) {
	Account* newAccount = new Account(0, name, freeaccount);
	freeaccount++;

	AccountElement* lnew = new AccountElement;
	lnew->account = newAccount;
	lnew->nextAccountElement = NULL;

	if (accountset==NULL) {
		accountset = lnew;
	} else {
		AccountElement* foo;
		for (foo = accountset; foo != NULL && foo->nextAccountElement != NULL; foo = foo->nextAccountElement) {
		}
		foo->nextAccountElement = lnew;
	}

	return freeaccount-1;
}

ResultFindAccount AccountsList::findAccount(AccountNumber n) {
	ResultFindAccount f;
	AccountElement* foo;

	for (foo = accountset; foo != NULL && foo->nextAccountElement != NULL && foo->account->getAccountNumber() != n; foo = foo->nextAccountElement) {
	}

	if (foo) {
		f.r = true;
		f.a = foo->account;
	} else {
		f.r = false;
	}
	return f;
}

#endif /* _ACCOUNTS_LIST_HPP */
