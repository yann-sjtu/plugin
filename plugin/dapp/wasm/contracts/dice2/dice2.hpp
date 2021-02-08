#include "dice2.pb.h"

#ifdef __cplusplus //而这一部分就是告诉编译器，如果定义了__cplusplus(即如果是cpp文件，
extern "C" { //因为cpp文件默认定义了该宏),则采用C语言方式进行编译
#endif

int startgame(int64_t amount);
int deposit(int64_t amount);
int play(int64_t amount, int64_t number);
int draw();
int stopgame();

#ifdef __cplusplus
}
#endif

//void withdraw(char* creator);
dice2::gamestatus get_status();
void set_status(dice2::gamestatus status);
//void update_addrinfo(char* addr, int64_t amount, int64_t earnings);
void set_roundinfo(dice2::roundinfo info);
dice2::roundinfo get_roundinfo(int64_t round);
void gen_roundinfo_key(char* round_key, int64_t round);
//bool is_active();
