
#include <cstdio>
#include <chrono>
#include <thread>

int main() {
  using namespace std::chrono_literals;
  auto get_clock = [] { return std::chrono::system_clock::system_clock::now(); };
  auto next = get_clock() + 1s;
  while (get_clock() < next);
  printf("hello world\n");
}
