using namespace std;

#include "Account.hpp"
#include "Accounts.hpp"

class AccountsList : public Accounts {
private:
	struct AccountElement {
		Account* account;
		AccountElement* nextAccountElement;
	};
	AccountElement* accountset;
	AccountNumber freeaccount; // required by constructor of Account...
public:
	AccountsList();
	~AccountsList();
	AccountNumber createAccount(string);
	ResultFindAccount findAccount(AccountNumber);
};
