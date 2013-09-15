#include "Account.hpp"

Account::Account(Euro b, string s, AccountNumber n) {
    money = b;
    name = s;
    number = n;
}

AccountNumber Account::getAccountNumber() {
    return number;
}

void Account::putMoney(Euro b) {
    money = money + b;
}

bool Account::takeMoney(Euro b) {
    if (money >= b) {
        money = money - b;
        return true;
    } else {
        return false;
    }
}

Euro Account::getBalance() {
    return money;
}
