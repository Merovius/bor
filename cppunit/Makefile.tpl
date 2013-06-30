LDFLAGS+=-lcppunit

%.o: %.cpp
	$(CXX) -c -o $@ $<
