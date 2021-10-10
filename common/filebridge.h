#ifndef __FILE_BRIDGE_H__
#define __FILE_BRIDGE_H__

#include <unistd.h>

ssize_t bridge_read(int fd, char *buf, size_t count);

ssize_t bridge_write(int fd, char *buf, size_t count);

void init_thread_pool(int io_threads, int prior_io_threads);

void destroy_thread_pool();

void bridge_pool_read(int *args);

void bridge_pool_write(int *args);

void bridge_pool_open(int *args);

void bridge_pool_rename(int *args);

void bridge_pool_close(int *args);

#endif//__FILE_BRIDGE_H__