package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "time"

    "go.uber.org/fx"
)

// NewLogger constructs a logger. It's just a regular Go function, without any
// special relationship to Fx.
//
// Since it returns a *log.Logger, Fx will treat NewLogger as the constructor
// function for the standard library's logger. (We'll see how to integrate
// NewLogger into an Fx application in the main function.) Since NewLogger
// doesn't have any parameters, Fx will infer that loggers don't depend on any
// other types - we can create them from thin air.
//
// Fx calls constructors lazily, so NewLogger will only be called only if some
// other function needs a logger. Once instantiated, the logger is cached and
// reused - within the application, it's effectively a singleton.
//
// By default, Fx applications only allow one constructor for each type. See
// the documentation of the In and Out types for ways around this restriction.
/*
	NewLogger 构造了一个logger,它只是常规的Go函数，与Fx没有任何特殊关系。

	由于返回的是* log.Logger，Fx将把NewLogger视为标准库的logger的构造函数。 （我们将
	了解如何集成由于NewLogger没有任何参数，因此Fx会推断出logger不依赖于任何其他类型-
	所以我们可以凭空创建它们。

	Fx调用构造函数是慵懒的，所以只有在某些其他函数需要logger时才调用NewLogger。 一旦实
	例化，logger便被缓存与复用-在应用程序内，它实际上是单例(设计模式的一种)。

	默认情况下，Fx应用程序仅允许每种类型使用一个构造函数。 有关此限制的解决方法，请参见
	输入和输出类型的文档。
*/
func NewLogger() *log.Logger {
    logger := log.New(os.Stdout, "" /* prefix */, 0 /* flags */)
    logger.Print("Executing NewLogger.")
    return logger
}

// NewHandler constructs a simple HTTP handler. Since it returns an
// http.Handler, Fx will treat NewHandler as the constructor for the
// http.Handler type.
//
// Like many Go functions, NewHandler also returns an error. If the error is
// non-nil, Go convention tells the caller to assume that NewHandler failed
// and the other returned values aren't safe to use. Fx understands this
// idiom, and assumes that any function whose last return value is an error
// follows this convention.
//
// Unlike NewLogger, NewHandler has formal parameters. Fx will interpret these
// parameters as dependencies: in order to construct an HTTP handler,
// NewHandler needs a logger. If the application has access to a *log.Logger
// constructor (like NewLogger above), it will use that constructor or its
// cached output and supply a logger to NewHandler. If the application doesn't
// know how to construct a logger and needs an HTTP handler, it will fail to
// start.
//
// Functions may also return multiple objects. For example, we could combine
// NewHandler and NewLogger into a single function:
//
//   func NewHandlerAndLogger() (*log.Logger, http.Handler, error)
//
// Fx also understands this idiom, and would treat NewHandlerAndLogger as the
// constructor for both the *log.Logger and http.Handler types. Just like
// constructors for a single type, NewHandlerAndLogger would be called at most
// once, and both the handler and the logger would be cached and reused as
// necessary.
/*
	NewHandler构造一个简单的HTTP handler。 由于返回了http.Handler，Fx将把NewHandler
	视为http.Handler类型的构造函数。

	像许多Go函数一样，NewHandler也返回错误。 如果err不为nil，则Go规定将告诉调用者假定
	为NewHandler失败，并且其他的返回值不能安全使用。 Fx理解了这个习惯用法，并假定最后一
	个返回值是err的任何函数都遵循此约定。

	与NewLogger不同，NewHandler具有形式参数。 Fx会将这些参数解释为依赖项：为了构造HTTP
	Handler，NewHandler需要logger。 如果应用程序可以访问* log.Logger构造函数（如上述的
	NewLogger），它将使用该构造函数或其缓存的输出并将logger提供给NewHandler。 如果应用程
	序不知道如何构造logger，并且需要HTTP处理程序，它将无法启动。

	函数也可能返回多个对象。 例如，我们可以将NewHandler和NewLogger组合成一个函数：
	  func NewHandlerAndLogger() (*log.Logger, http.Handler, error)

	Fx也理解这个习惯用法，并将NewHandlerAndLogger视为*log.Logger和http.Handler类型的
	构造函数。 就像单一类型的构造函数，NewHandlerAndLogger最多将被调用一次，Handler和
	logger都将被缓存并根据需要重新使用。
*/
func NewHandler(logger *log.Logger) (http.Handler, error) {
    logger.Print("Executing NewHandler.")
    return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
        logger.Print("Got a request.")
    }), nil
}

// NewMux constructs an HTTP mux. Like NewHandler, it depends on *log.Logger.
// However, it also depends on the Fx-specific Lifecycle interface.
//
// A Lifecycle is available in every Fx application. It lets objects hook into
// the application's start and stop phases. In a non-Fx application, the main
// function often includes blocks like this:
//
//   srv, err := NewServer() // some long-running network server
//   if err != nil {
//     log.Fatalf("failed to construct server: %v", err)
//   }
//   // Construct other objects as necessary.
//   go srv.Start()
//   defer srv.Stop()
//
// In this example, the programmer explicitly constructs a bunch of objects,
// crashing the program if any of the constructors encounter unrecoverable
// errors. Once all the objects are constructed, we start any background
// goroutines and defer cleanup functions.
//
// Fx removes the manual object construction with dependency injection. It
// replaces the inline goroutine spawning and deferred cleanups with the
// Lifecycle type.
//
// Here, NewMux makes an HTTP mux available to other functions. Since
// constructors are called lazily, we know that NewMux won't be called unless
// some other function wants to register a handler. This makes it easy to use
// Fx's Lifecycle to start an HTTP server only if we have handlers registered.
/*
	NewMux构造一个HTTP mux。 与NewHandler一样，它依赖 *log.Logger，但是也依赖Fx
	特定的Lifecycle接口。

	每个Fx应用程序都有一个生命周期。 它使对象可以hook进入应用程序的开始和停止阶段。 在
	非Fx应用程序中，main function 通常包括以下块：

		srv, err := NewServer() // 一些长期运行的网络服务器
		if err != nil {
			log.Fatalf("failed to construct server: %v", err)
		}
		// 根据需要构造其他对象
		go srv.Start()
		defer srv.Stop()
	  
	在这个例子中，程序员显式地构造了一堆对象，如果任何构造函数遇到不可恢复的错误，程序
	就会崩溃。 构造完所有对象后，我们将启动任何后台goroutine并defer清除功能。

	Fx删除了依赖注入的手动对象构造。 用Lifecycle类型替换了inline goroutine生成和延时
	清除。

	在这里，NewMux使HTTP mux可用于其他功能。 由于构造函数是延迟调用的，因此我们知道除
	非有其他函数想要注册的行为，否则不会调用NewMux。 只有当我们已注册处理程序时，这让
	使用Fx的LifeSycle启动HTTP服务器变得容易。
*/
func NewMux(lc fx.Lifecycle, logger *log.Logger) *http.ServeMux {
    logger.Print("Executing NewMux.")
    // First, we construct the mux and server. We don't want to start the server
	// until all handlers are registered.
	// 首先，我们构建mux和server。 在所有处理程序都注册之前，我们不希望启动服务器。
    mux := http.NewServeMux()
    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }
    // If NewMux is called, we know that another function is using the mux. In
    // that case, we'll use the Lifecycle type to register a Hook that starts
    // and stops our HTTP server.
    //
    // Hooks are executed in dependency order. At startup, NewLogger's hooks run
    // before NewMux's. On shutdown, the order is reversed.
    //
    // Returning an error from OnStart hooks interrupts application startup. Fx
    // immediately runs the OnStop portions of any successfully-executed OnStart
    // hooks (so that types which started cleanly can also shut down cleanly),
    // then exits.
    //
    // Returning an error from OnStop hooks logs a warning, but Fx continues to
	// run the remaining hooks.
	/*
		如果NewMux被调用，我们知道另一个函数正在使用这个mux。 这种情况下，我们将使用
		lifecycle类型注册一个用于启动和停止HTTP服务器的Hook。

		hooks 按依赖关系顺序执行。 在启动时，NewLogger的hooks先于NewMux的hooks运行。 
		关机时，顺序相反。
		
		从OnStart hooks 返回错误会中断应用程序启动。 Fx立即运行任何成功执行的OnStart
		hooks的OnStop部分（这样，干净启动的类型也可以干净关闭），然后退出。

		从OnStop hooks 返回错误会记录警告，但是Fx继续运行其余的挂钩。
	*/
    lc.Append(fx.Hook{
        // To mitigate the impact of deadlocks in application startup and
        // shutdown, Fx imposes a time limit on OnStart and OnStop hooks. By
        // default, hooks have a total of 15 seconds to complete. Timeouts are
		// passed via Go's usual context.Context.
		/*
		为了减轻死锁对应用程序启动和关闭的影响，Fx对OnStart和OnStop hooks 施加了时间限制。
		默认情况下，挂钩总共需要15秒才能完成。 超时是通过Go的常规context.Context传递的。
		*/
        OnStart: func(context.Context) error {
            logger.Print("Starting HTTP server.")
            // In production, we'd want to separate the Listen and Serve phases for
			// better error-handling.
			// 在生产中，我们希望将Listen和Server阶段分开以更好地处理错误。
            go server.ListenAndServe()
            return nil
        },
        OnStop: func(ctx context.Context) error {
            logger.Print("Stopping HTTP server.")
            return server.Shutdown(ctx)
        },
    })

    return mux
}

// Register mounts our HTTP handler on the mux.
//
// Register is a typical top-level application function: it takes a generic
// type like ServeMux, which typically comes from a third-party library, and
// introduces it to a type that contains our application logic. In this case,
// that introduction consists of registering an HTTP handler. Other typical
// examples include registering RPC procedures and starting queue consumers.
//
// Fx calls these functions invocations, and they're treated differently from
// the constructor functions above. Their arguments are still supplied via
// dependency injection and they may still return an error to indicate
// failure, but any other return values are ignored.
//
// Unlike constructors, invocations are called eagerly. See the main function
// below for details.
/*
	Register函数将我们的HTTP hander挂载在mux上。 

	Register是典型的顶级应用程序函数：它采用了ServeMux之类的通用类型，该类型通常来
	自第三方库，并将其引入包含我们的应用程序逻辑的类型。 在这种情况下，该介绍包括注册
	HTTP处理程序。 其他典型示例包括注册RPC过程和启动队列使用者。

	Fx调用这些函数调用，并且它们与上述构造函数的区别对待。 它们的参数仍通过依赖项注入
	提供，并且它们仍可能返回错误以指示失败，但是任何其他返回值都将被忽略。

	与构造函数不同，invocations 被急切地调用。 有关详细信息，请参见下面的主要功能。

*/
func Register(mux *http.ServeMux, h http.Handler) {
    mux.Handle("/", h)
}

func main() {
    app := fx.New(
        // Provide all the constructors we need, which teaches Fx how we'd like to
        // construct the *log.Logger, http.Handler, and *http.ServeMux types.
        // Remember that constructors are called lazily, so this block doesn't do
	// much on its own.
		/*
		提供我们需要的所有构造函数，这将教给Fx我们如何构造* log.Logger，http.Handler和
		* http.ServeMux类型。请记住，构造函数被懒惰地调用，因此，该块本身并不会做太多事情。
		*/
        fx.Provide(
            NewLogger,
            NewHandler,
            NewMux,
        ),
        // Since constructors are called lazily, we need some invocations to
        // kick-start our application. In this case, we'll use Register. Since it
        // depends on an http.Handler and *http.ServeMux, calling it requires Fx
        // to build those types using the constructors above. Since we call
        // NewMux, we also register Lifecycle hooks to start and stop an HTTP
		// server.
		/*
		由于构造函数是延迟调用的，因此我们需要一些invocations才能启动我们的应用程序。 在
		这种情况下，我们将使用Register。 由于它依赖于http.Handler和* http.ServeMux，
		因此调用它需要Fx使用上面的构造函数来构建这些类型。 由于我们称为NewMux，因此我们还
		注册了Lifecycle挂钩来启动和停止HTTP服务器。
		*/
        fx.Invoke(Register),
    )

    // In a typical application, we could just use app.Run() here. Since we
    // don't want this example to run forever, we'll use the more-explicit Start
	// and Stop.
	/*
	在典型的应用程序中，我们可以在此处使用app.Run()。 由于我们不希望该示例永远运行，
	因此我们将使用更加明确的Start和Stop。
	*/
    startCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    if err := app.Start(startCtx); err != nil {
        log.Fatal(err)
    }

    // Normally, we'd block here with <-app.Done(). Instead, we'll make an HTTP
	// request to demonstrate that our server is running.
	/*
	通常，我们在这里使用<-app.Done（）进行阻止。 相反，我们将发出HTTP请求以证明我们的服务器正在运行。
	*/
    http.Get("http://localhost:8080/")

    stopCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()
    if err := app.Stop(stopCtx); err != nil {
        log.Fatal(err)
    }

}
