
#include <cstdio>
#include <chrono>
#include <thread>

int main() {
  using namespace std::chrono_literals;
  std::this_thread::sleep_until(std::chrono::system_clock::system_clock::now() + 1s);
  printf("hello world\n");
}
