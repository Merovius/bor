#include "Bank.hpp"

Bank::Bank(string n, string m) {
	name = n;
	mode = m;

	if (mode=="array") {
		accounts = new AccountsArray();
	} else {
		accounts = new AccountsList();
	}
}

/**
 * Create an account and decrease the amount of available accounts
 * Return the account number of created account
 */
AccountNumber Bank::createAccount(string name) {
	int freeaccount = accounts->createAccount(name);

	return freeaccount;
}

/**
 * Search for an accont and either take money away (return amount) or
 * return 0 when account not found
 */
Euro Bank::takeMoney(AccountNumber n, Euro b) {

	ResultFindAccount foundAccount = accounts->findAccount(n);

	if (foundAccount.r) {
		bool ok = foundAccount.a->takeMoney(b);
		if (ok)
			return b;
		else {
			return 0;
		};
	} else {
		return 0;
	}
}

/**
 * Search for an accont and execute either the money transfer (return true) or
 * return false when account not found
 */
bool Bank::transferMoney(AccountNumber n1, Euro e, Bank* b, AccountNumber n2) {
	bool ok1=true, ok2=true, ok3=true;

	ResultFindAccount foundAccount = accounts->findAccount(n1);

	ok1 = foundAccount.r;
	Account* account = foundAccount.a;
	if (ok1) {
		ok2 = account->takeMoney(e);
		if (ok2) {
			ok3 = b->putMoney(n2,e);
			if (!ok3)
				foundAccount.a->putMoney(e);
		}
	} else {
	}

	return ok1;
}

/**
 * Search for an accont and either put money on this account (return true)
 * or return false when account not found
 */
bool Bank::putMoney(AccountNumber n, Euro b) {
	ResultFindAccount foundAccount = accounts->findAccount(n);

	if (foundAccount.r) {
		foundAccount.a->putMoney(b);
		return true;
	} else {
		return false;
	}
}

/**
 * Search for an account and either return its balance, or 0
 */
Euro Bank::balance(AccountNumber n) {
	ResultFindAccount found = accounts->findAccount(n);

	if (found.r) {
        return found.a->getBalance();
	}
    return 0;
}
