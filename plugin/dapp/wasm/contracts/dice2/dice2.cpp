#include "../common.h"
#include "dice2.hpp"
#include <string.h>

#define STATUS "dice_status\0"
#define ROUND_KEY_PREFIX "roundinfo:"
#define ROUND_KEY_PREFIX_LEN 10
#define OK 0

int startgame(int64_t amount) {
    char from[34]={0};
    getFrom(from, 34);

    printlog(from, 34);
    dice2::gamestatus status = get_status();
    if (status.is_active()) {
        const char info[] = "active game\0";
        printlog(info, string_size(info));
        return -1;
    }
    if ((status.height() != 0) && (strncmp(from, status.game_creator().c_str(), 34) != 0)) {
        const char info[] = "game can only be restarted by the creator\0";
        printlog(info, string_size(info));
        return -1;
    }
    if (amount <= 0) {
       return -1;
    }

    if (OK != execFrozen(from, 34, amount)) {
        const char info[] = "frozen coins failed\0";
        printlog(info, string_size(info));
        return -1;
    }
    status.set_height(getHeight());
    status.set_is_active(true);
    status.set_deposit(amount);
    status.set_game_creator(from);
    status.set_game_balance(amount);
    set_status(status);
    const char info[] = "call contract success\0";
    printlog(info, string_size(info));
    return 0;
}

int deposit(int64_t amount) {
    dice2::gamestatus status = get_status();
    set_status(status);
    if (!status.is_active()) {
        const char info[] = "inactive game\0";
        printlog(info, string_size(info));
        return -1;
    }
    char from[34]={0};
    getFrom(from, 34);
    printlog(from, 34);
    printlog(status.game_creator().c_str(), 34);
    if (strncmp(from, status.game_creator().c_str(), 34) != 0) {
        const char info[] = "game can only be deposited by the creator\0";
        printlog(info, string_size(info));
        return -1;
    }
    if (amount<=0) {
        return -1;
    }
    if (OK != execFrozen(from, 34, amount)) {
        const char info[] = "frozen coins failed\0";
        printlog(info, string_size(info));
        return -1;
    }
    status.set_deposit(status.deposit() + amount);
    status.set_game_balance(status.game_balance() + amount);
    set_status(status);
    return 0;
}

int play(int64_t amount, int64_t number) {
    dice2::gamestatus status = get_status();
    if (!status.is_active()) {
        const char info[] = "inactive game\0";
        printlog(info, string_size(info));
        return -1;
    }
    if (number<2 || number>97) {
        const char info[] = "number must be within range of [2,97]\0";
        printlog(info, string_size(info));
        return -1;
    }
    if (amount<=0) {
        return -1;
    }
    //最大投注额为奖池的0.5%
    if (amount*200>status.game_balance()) {
        const char info[] = "amount is too big\0";
        printlog(info, string_size(info));
        return -1;
    }
    char from[34]={0};
    getFrom(from, 34);
    if (OK != execFrozen(from, 34, amount)) {
        const char info[] = "frozen coins failed\0";
        printlog(info, string_size(info));
        return -1;
    }
    status.set_current_round(status.current_round() + 1);
    status.set_total_bets(status.total_bets() + amount);
    set_status(status);

    dice2::roundinfo info;
    info.set_round(status.current_round());
    info.set_height(getHeight());
    info.set_player(from);
    info.set_amount(amount);
    info.set_guess_num(number);
    set_roundinfo(info);
    return 0;
}

int draw() {
    dice2::gamestatus status = get_status();
    if (!status.is_active()) {
        const char info[] = "inactive game\0";
        printlog(info, string_size(info));
        return -1;
    }
    char from[34]={0};
    getFrom(from, 34);
    if (strncmp(from, status.game_creator().c_str(), 34) != 0) {
        const char info[] = "game can only be drawn by the creator\0";
        printlog(info, string_size(info));
        return -1;
    }
    if (status.current_round() == status.finished_round()) {
        //没有待开奖记录
        return 0;
    }

    int64_t height = getHeight();
    status.set_height(height);
    int64_t random = getRandom() % 100;
    printint(random);
    for (int64_t round=status.finished_round()+1;round<=status.current_round();round++) {
        dice2::roundinfo info = get_roundinfo(round);
        if (info.height() == status.height()) {
            break;
        }
        if (random < info.guess_num()) {
            int64_t probability = info.guess_num();
            int64_t payout = info.amount() *(100 - probability) / probability;
            if (OK != execTransferFrozen(status.game_creator().c_str(), 34, info.player().c_str(), 34, payout-info.amount())) {
                const char info[] = "transfer frozen coins from game creator failed\0";
                printlog(info, string_size(info));
                return -1;
            }
            if (OK != execActive(info.player().c_str(), 34, info.amount())) {
                const char info[] = "active frozen coins failed\0";
                printlog(info, string_size(info));
                return -1;
            }

            status.set_total_player_win(status.total_player_win()+ payout);
            status.set_game_balance(status.game_balance() + info.amount() - payout);
            info.set_player_win(true);

        } else {
            if (OK != execTransferFrozen(info.player().c_str(), 34, status.game_creator().c_str(), 34, info.amount())) {
                const char info[] = "transfer frozen coins from player failed\0";
                printlog(info, string_size(info));
                return -1;
            }
            if (OK != execFrozen(status.game_creator().c_str(), 34, info.amount())) {
                const char info[] = "frozen coins failed\0";
                printlog(info, string_size(info));
                return -1;
            }
            info.set_player_win(false);
            status.set_game_balance(status.game_balance() + info.amount());
        }

        info.set_rand_num(random);
        info.set_is_finished(true);
        status.set_finished_round(status.finished_round()+1);
        set_roundinfo(info);
        set_status(status);
    }

    return 0;
}

int stopgame() {
    dice2::gamestatus status = get_status();
    char from[34]={0};
    getFrom(from, 34);
    if (strncmp(from, status.game_creator().c_str(), 34) != 0) {
        const char info[] = "game can only be stopped by the creator\0";
        printlog(info, string_size(info));
        return -1;
    }

    if (status.finished_round() != status.current_round()) {
        // const char info[] = "inactive game\0";
        // printlog(info, string_size(info));
        return -1;
    }

    if (!status.is_active()) {
        const char info[] = "inactive game\0";
        printlog(info, string_size(info));
        return -1;
    }
    if (OK != execActive(from, 34, status.game_balance())) {
        const char info[] = "active frozen coins failed\0";
        printlog(info, string_size(info));
        return -1;
    }
    status.clear_is_active();
    status.clear_deposit();
    status.clear_game_balance();
    set_status(status);
    return 0;
}



dice2::gamestatus get_status() {
    char status_key[] = STATUS;
    size_t size = getStateDBSize(status_key, string_size(status_key));
    char *buf = (char *)malloc( sizeof(char)*size );
    getStateDB(status_key, string_size(status_key), buf, size);
    dice2::gamestatus status{};
    status.ParseFromArray(buf, size);
    return status;
}

void set_status(dice2::gamestatus status) {
    char status_key[] = STATUS;
    size_t size = status.ByteSizeLong();
    char *buf = (char *)malloc( sizeof(char)*size );
    status.SerializeToArray(buf, size);
    setStateDB(status_key, string_size(status_key), buf, size);
}

dice2::roundinfo get_roundinfo(int64_t round) {
    char round_key[32];
    gen_roundinfo_key(round_key, round);
    size_t size = getStateDBSize(round_key, string_size(round_key));
    char *buf = (char *)malloc( sizeof(char)*size );
    getStateDB(round_key, string_size(round_key), buf, size);
    dice2::roundinfo info;
    info.ParseFromArray(buf, size);
    return info;
}

void set_roundinfo(dice2::roundinfo info) {
    char round_key[32];
    gen_roundinfo_key(round_key, info.round());
    size_t size = info.ByteSizeLong();
    char *buf = (char *)malloc( sizeof(char)*size );
    info.SerializeToArray(buf, size);
    setStateDB(round_key, string_size(round_key), buf, size);
}

void gen_roundinfo_key(char* round_key, int64_t round) {
    strcpy(round_key, ROUND_KEY_PREFIX);
    char round_str[20] = {0};
    int index;
    for (index=0;;index++) {
        round_str[index] = char(round%10) + '0';
        round/=10;
        if (round==0) {
            break;
        }
    }
    for (int i=0;i<=index;i++) {
        round_key[ROUND_KEY_PREFIX_LEN+i] = round_str[index-i];
    }
    round_key[ROUND_KEY_PREFIX_LEN+index+1] = '\0';
}

