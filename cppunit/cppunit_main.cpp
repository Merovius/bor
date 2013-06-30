#include <iostream>
#include <cppunit/Exception.h>
#include <cppunit/extensions/TestFactoryRegistry.h>
#include <cppunit/Message.h>
#include <cppunit/SourceLine.h>
#include <cppunit/Test.h>
#include <cppunit/TestFailure.h>
#include <cppunit/TestRunner.h>
#include <cppunit/TestResult.h>
#include <cppunit/TestResultCollector.h>
#include <cppunit/TestSuite.h>

CppUnit::Test *global_suite;

class TAPListener : public CppUnit::TestListener {
    public:
        void startSuite(CppUnit::Test *suite);
        void startTest(CppUnit::Test *test);
        void addFailure(const CppUnit::TestFailure &failure);
        void endTest(CppUnit::Test *test);

    private:
        CppUnit::Message msg;
        bool success;
};

void TAPListener::startSuite(CppUnit::Test *suite) {
    if (global_suite == suite)
        return;
    CppUnit::TestSuite *s = (CppUnit::TestSuite *)suite;
    std::cout << "TAP version 13" << std::endl;
    std::cout << "1.." << s->getChildTestCount() << std::endl;
}

void TAPListener::startTest(CppUnit::Test *test) {
    success = true;
}

void TAPListener::addFailure(const CppUnit::TestFailure &failure) {
    CppUnit::SourceLine source = failure.sourceLine();
    CppUnit::Exception *exp = failure.thrownException();
    msg = exp->message();

    success = false;
}

void TAPListener::endTest(CppUnit::Test *test) {
    if (success) {
        std::cout << "ok " << test->getName() << std::endl;
        return;
    }
    std::cout << "not ok " << test->getName() << std::endl;
    std::cout << "# " << msg.shortDescription() << std::endl;
    for (int i = 0; i < msg.detailCount(); i++) {
        std::cout << "# \t" << msg.detailAt(i) << std::endl;
    }
}


int main(int argc, char* argv[]) {
    // Get the top level suite from the registry
    global_suite = CppUnit::TestFactoryRegistry::getRegistry().makeTest();

    // Create the event manager and test controller
    CppUnit::TestResult controller;

    // Add a listener that colllects test result
    TAPListener listener;
    controller.addListener(&listener);

    // Adds the test to the list of test to run
    CppUnit::TestRunner runner;
    runner.addTest(global_suite);

    // Run the tests.
    const std::string path = "";
    runner.run(controller, path);

    return 0;
}
