#include "filebridge.h"

#include <pthread.h>
#include <ctype.h>
#include <errno.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>


typedef struct tpool_work {
	void*              (*work_routine)(void*); //function to be called
	void*              args;                   //arguments 
	struct tpool_work* next;
} tpool_work_t;
 
typedef struct tpool {
	size_t               shutdown;       //is tpool shutdown or not, 1 ---> yes; 0 ---> no
	size_t               maxnum_thread;  // maximum of threads
	pthread_t            *thread_id;     // a array of threads
	tpool_work_t*        tpool_head;     // tpool_work queue
	pthread_cond_t       queue_ready;    // condition varaible
	pthread_mutex_t      queue_lock;     // queue lock
} tpool_t;

static void* thread_routine(void *args)
{
	tpool_t* pool = (tpool_t*)args;
	tpool_work_t* work = NULL;

	while (1) {
		pthread_mutex_lock(&pool->queue_lock);
		while (!pool->tpool_head && !pool->shutdown) {
			// If there is no works and pool is not shutdown, it should be suspended for being awake
			pthread_cond_wait(&pool->queue_ready,&pool->queue_lock);
		}

		if (pool->shutdown) {
			pthread_mutex_unlock(&pool->queue_lock);//pool shutdown,release the mutex and exit
			pthread_exit(NULL);
		}

		/* tweak a work*/
		work = pool->tpool_head;
		pool->tpool_head = (tpool_work_t*)pool->tpool_head->next;
		pthread_mutex_unlock(&pool->queue_lock);

		work->work_routine(work->args);

		free(work);
	}
	return NULL;
}

int create_tpool(tpool_t **pool, size_t max_thread_num, size_t prior_io_threads)
{
	(*pool) = (tpool_t*)malloc(sizeof(tpool_t));
	if (NULL == *pool) {
		dprintf(1, "In %s, malloc tpool_t failed! errno = %d, explain: %s\n", __func__, errno, strerror(errno));
		exit(-1);
	}

	(*pool)->shutdown = 0;
	(*pool)->maxnum_thread = max_thread_num;
	(*pool)->thread_id = (pthread_t*)malloc(sizeof(pthread_t)*max_thread_num);
	if ((*pool)->thread_id == NULL) {
		dprintf(1, "In %s, init thread id failed, errno = %d, explain: %s", __func__, errno, strerror(errno));
		exit(-1);
	}

	(*pool)->tpool_head = NULL;
	if (pthread_mutex_init(&((*pool)->queue_lock), NULL) != 0) {
		dprintf(1, "In %s, initial mutex failed,errno = %d, explain: %s", __func__, errno, strerror(errno));
		exit(-1);
	}
	
	if (pthread_cond_init(&((*pool)->queue_ready), NULL) != 0) {
		dprintf(1, "In %s,initial condition variable failed, errno = %d, explain: %s", __func__, errno, strerror(errno));
		exit(-1);
	}

	pthread_attr_t attr;
	struct sched_param sched;
	int rs = pthread_attr_init(&attr);
	if (rs != 0) {
		dprintf(1, "Init thread attr failed.\n");
		perror("pthread_attr_init");
		exit(-1);
	}

	struct sched_param param;
	param.sched_priority = 51;
	pthread_attr_setschedpolicy(&attr, SCHED_RR);
	pthread_attr_setschedparam(&attr, &param);
	pthread_attr_setinheritsched(&attr, PTHREAD_EXPLICIT_SCHED);//要使优先级其作用必须要有这句话
	
	for (int i = 0; i < max_thread_num; i++) {
		pthread_attr_t *at = i < prior_io_threads ? &attr : NULL;
		if (pthread_create(&((*pool)->thread_id[i]), at, thread_routine, (void*)(*pool)) != 0) {
			printf("pthread_create failed!\n");
			if (i < prior_io_threads) {
				dprintf(1, "You had set prior io threads, so try run it with sudo.\n");
			}
			exit(-1);
		}
	}
	return 0;
}
 
void destroy_tpool(tpool_t *pool)
{
	tpool_work_t* tmp_work;
	
	if (pool->shutdown) {
		return;
	}
	pool->shutdown = 1;
	
	pthread_mutex_lock(&pool->queue_lock);
	pthread_cond_broadcast(&pool->queue_ready);
	pthread_mutex_unlock(&pool->queue_lock);
	
	for (int i = 0; i < pool->maxnum_thread; i++) {
		pthread_join(pool->thread_id[i], NULL);
	}

	free(pool->thread_id);
	while (pool->tpool_head) {
		tmp_work = pool->tpool_head;
		pool->tpool_head = (tpool_work_t*)pool->tpool_head->next;
		free(tmp_work);
	}
	
	pthread_mutex_destroy(&pool->queue_lock);
	pthread_cond_destroy(&pool->queue_ready);
	free(pool);
}
 
int add_task_2_tpool(tpool_t *pool, void* (*routine)(void*), void *args)
{
	struct tpool_work* work, *member;
	
	if (!routine) {
		printf("rontine is null!\n");
		return -1;
	}
	
	work = (struct tpool_work*)malloc(sizeof(struct tpool_work));
	if (!work) {
		return -1;
	}
	
	work->work_routine = routine;
	work->args = args;
	work->next = NULL;
	
	pthread_mutex_lock(&pool->queue_lock);
	member = pool->tpool_head;
	if (!member) {
		pool->tpool_head = work;
	} else {
		while (member->next) {
			member = (struct tpool_work*)member->next;
		}
		member->next = work;
	}
	
	//notify the pool that new task arrived!
	pthread_cond_signal(&pool->queue_ready);
	pthread_mutex_unlock(&pool->queue_lock);
	return 0;
}

static tpool_t* pool = NULL;

void init_thread_pool(int io_threads, int prior_io_threads)
{
	if (pool != NULL) {
		return;
	} else if (io_threads < 1) {
		dprintf(1, "Io thread must > 0.\n");
		return;
	}

	if (0 != create_tpool(&pool, io_threads, prior_io_threads)) {
		dprintf(1, "create_tpool failed!\n");
		return;
	}
}

void destroy_thread_pool()
{
	if (pool == NULL) {
		return;
	}

	destroy_tpool(pool);
}

ssize_t bridge_read(int fd, char *buf, size_t count)
{
	return read(fd, buf, count);
}

ssize_t bridge_write(int fd, char *buf, size_t count)
{
	return write(fd, (const void *)buf, count);
}

struct GoArgs {
	int fd;
	int n;
	int err;
	int cap;
	char *buf;
};

void go_done_callback(int*);
void go_done_open_callback(int*);
void go_done_close_callback(int*);
void go_done_rename_callback(int*);
void go_debug_log(char*);

void* FuncRead(void* args)
{
	struct GoArgs *ga = (struct GoArgs *)args;
	size_t n = read(ga->fd, ga->buf, ga->cap);
	if (n < 1) {
		ga->err = errno;
	} else {
		ga->err = 0;
	}

	ga->n = n;
	go_done_callback((int *)ga);

	go_debug_log("Write done...............");
	return NULL;
}

void bridge_pool_read(int *args)
{
	add_task_2_tpool(pool, FuncRead, (void*)args);
}

void* FuncWrite(void* args)
{
	// go_debug_log("in FuncWrite..............");
	struct GoArgs *ga = (struct GoArgs *)args;

	// go_debug_log(ga->buf);
	// dprintf(STDOUT_FILENO, "get data======fd: %d, len: %d=========%s\n", ga->fd, ga->cap, ga->buf);
	size_t n = write(ga->fd, ga->buf, ga->cap);
	// go_debug_log("write done...............");
	if (n < 1) {
		ga->err = errno;
	} else {
		ga->err = 0;
	}

	ga->n = n;
	go_done_callback((int *)ga);

	// go_debug_log("Write done...............");
	return NULL;
}

void bridge_pool_write(int *args)
{
	// go_debug_log("in bridge_pool_write.");
	add_task_2_tpool(pool, FuncWrite, (void*)args);
}

struct GoOpenArgs {
	int fd;
	int flag;
	int mode;
	char *path;
	int err;
};

void* FuncOpen(void* args)
{
	// go_debug_log("in FuncOpen..............");
	struct GoOpenArgs *ga = (struct GoOpenArgs *)args;

	int fd = open(ga->path, ga->flag, ga->mode);
	if (fd < 1) {
		ga->err = errno;
		// dprintf(1, "open file err, fd: %d, errno: %d\n", fd, errno);
	} else {
		ga->err = 0;
		// dprintf(1, "open file success, fd: %d\n", fd);
	}

	ga->fd = fd;
	go_done_open_callback((int *)ga);

	return NULL;
}

void bridge_pool_open(int *args)
{
	// go_debug_log("in bridge_pool_open.");
	add_task_2_tpool(pool, FuncOpen, (void*)args);
}

struct GoCloseArgs {
	int ret;
	int fd;
	int err;
};

void* FuncClose(void* args)
{
	// go_debug_log("in FuncClose..............");
	struct GoCloseArgs *ga = (struct GoCloseArgs *)args;

	int ret = close(ga->fd);
	if (ret != 0) {
		ga->err = errno;
		// dprintf(1, "open file err, fd: %d, errno: %d\n", ga->fd, errno);
	} else {
		ga->err = 0;
	}

	ga->ret = ret;
	go_done_close_callback((int *)ga);
	return NULL;
}

void bridge_pool_close(int *args)
{
	// go_debug_log("in bridge_pool_close.");
	add_task_2_tpool(pool, FuncClose, (void*)args);
}

struct GoRenameArgs {
	int ret;
	char *oldname;
	char *newname;
	int err;
};

void* FuncRename(void* args)
{
	// go_debug_log("in FuncRename..............");
	struct GoRenameArgs *ga = (struct GoRenameArgs *)args;

	int ret = rename(ga->oldname, ga->newname);
	if (ret != 0) {
		ga->err = errno;
		// dprintf(1, "rename file err, old: %s, new: %s, errno: %d\n", ga->oldname, ga->newname, errno);
	} else {
		ga->err = 0;
	}

	ga->ret = ret;
	go_done_rename_callback((int *)ga);

	return NULL;
}

void bridge_pool_rename(int *args)
{
	// go_debug_log("in bridge_pool_rename.");
	add_task_2_tpool(pool, FuncRename, (void*)args);
}