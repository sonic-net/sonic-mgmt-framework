// Host Account Management
#ifndef TIMER_H
#define TIMER_H

#include <glib.h>   // GSourceFunc

/** Prototype for functions passed to gtimer_c::gtimer_c().
 *
 *  This is the signature of the user provided callback function
 *  that should be called upon expiration of a timer.
 *
 *  @param user_data_p Data passed to the callback function when
 *                     timer expires.
 *
 *  @return %false if the timer should be stopped. %true is
 *          timer should keep firing at regular interval. */

class gtimer_c
{
public:
    gtimer_c(double interval_sec, bool (* user_cback_p)(void * user_data_p), void * user_data_p, int priority=G_PRIORITY_DEFAULT);
    ~gtimer_c();

    void  stop();
    /** @brief Stop timer
     */

    void  start(double new_interval_sec=-1.0, void * user_data_p=NULL);
    /** @brief Start (or restart) timer.  If timer is already running it will
     *         be re-started with its original timeout.
     *
     *         This method optionally allows you to change the default timeout
     *         by specifying new_timeout.
     *
     *  @param new_interval_sec Timeout interval in seconds.
     *  @param user_data_p Data passed to the callback function when
     *                     timer expires.
     */

    void  conditinal_start(int new_interval_s=-1.0, void * user_data_p=NULL)
    {
        if (!active())
            start(new_interval_s, user_data_p);
    }
    /** @brief Start timer only if it is not already running. Otherwise, leave
     *         timer alone.
     *
     *         This method optionally allows you to change the default timeout
     *         by specifying new_timeout.
     */

    void  clear();
    /** @brief Make a timer expire immediately.  The reactor will immediately
     *         queue up the callback for execution.
     */

    void  set_timeout(double new_interval_sec);
    /** @brief Set the timeout.  This allows you to change the default
     *         timeout. An already running timer will continue with its old
     *         timeout unless start() is called again after setting new
     *         timeout.
     *
     *  @param new_interval_sec Timeout interval in seconds.
     */

    void  set_user_data(void * user_data_p);
    /** @brief Set callback arguments.  This allows you to change the default
     *         cb_args.  An already running timer will continue with its old
     *         cb_args unless start() is called again after setting new
     *         cb_args.
     */

    double time_remaining() const;
    /** @brief  Return how much time is remaining on a timer before it expires.

        @return Time remaining on a timer before it expires.
                floating-point number in seconds
     */

    bool  active() const      { return 0 != tid_m; }
    /** @brief Check whether a timer is currently running
     *
     *  @return True if timer is running, false otherwise.
     */

    double get_interval_s() const { return interval_sec_m; }
    /** @brief Get the timeout interval.
     *
     *  @return The timer interval as floating-point number in
     *          seconds.
     */

private:

    typedef GSource * (* new_source_func_pt) (guint interval);

    GSource             * source_pm;
    guint                 tid_m;

    double                interval_sec_m;
    guint                 interval_m;
    bool               (* user_cback_pm)(void * user_data_p);
    void                * user_data_pm;
    int                   priority_m;

    new_source_func_pt    g_timeout_source_new_funct_pm;

    static gboolean callback(gpointer user_data_p);  // Must match GSourceFunc signature.
};

#include <ostream>      // std::ostream
std::ostream & operator<<(std::ostream  & stream_r, const gtimer_c  & timer_r);

#endif /* TIMER_H */
