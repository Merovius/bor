LDFLAGS+=-lcppunit

%.o: %.cpp
	$(CXX) $(CXXFLAGS) -c -o $@ $<
