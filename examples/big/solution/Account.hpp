#ifndef _ACCOUNT_HPP
#define _ACCOUNT_HPP

#include <string>

using namespace std;

typedef int AccountNumber;
typedef int Euro;

class Account {
private:
    Euro money;
    string name;
    AccountNumber number;
public:
    Account(Euro b = 0, string s = "", AccountNumber n = 0);
    AccountNumber getAccountNumber();
    void putMoney(Euro b);
    bool takeMoney(Euro b);
    Euro getBalance() ;
};

#endif /* _ACCOUNT_HPP */
