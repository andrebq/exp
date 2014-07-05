;(function(){

    function Pipeline() {
        if (!(this instanceof Pipeline)) {
            return new Pipeline();
        }
        this.head = null;
        this.tail = null;
    }

    Pipeline.prototype.add = function(proc) {
        if (!(proc instanceof Process)) {
            proc = new Process(proc);
        }
        if (this.head === null) {
            this.head = proc;
        }
        if (this.tail !== null) {
            this.tail.connectToStdout(proc.$input);
        }
        this.tail = proc;
    };

    function Input(data) {
        this.$data = data;
    }

    Input.prototype.read = function() {
        if (this.valid()) {
            return this.$data.Data;
        } else {
            return;
        }
    };

    Input.prototype.valid = function() {
        return !this.$data.Closed;
    };

    Input.prototype.error = function() {
        if (this.$data.Error !== undefined) {
            return this.$data.Error;
        } else {
            return;
        }
    }

    function Output(subject) {
        this.$subject = subject;
    }

    Output.prototype.write = function(value) {
        this.$subject.onNext(value);
    }

    Output.prototype.abort = function(error) {
        this.$subject.onError(error);
    }

    Output.prototype.close = function() {
        this.$subject.onCompleted();
    }

    function defaultCallbackDone() {
        this.output.close();
    }

    function defaultCallbackError(error) {
        this.output.abort(error);
    }

    // Callback is any function that accept two parameters the (input) and
    // one (output).
    //
    // Every time the function is called, input describes the actual state of the process input
    // which could be: Valid (the given input is valid), Closed (no more data will be sent on this channel) and a 
    // optional flag indicating an error (Error).
    //
    // The callback MUST return a value describing the state of the output channel
    function Process(callback, callbackError, callbackDone) {
        if (!(this instanceof Process)){
            return new Process(callback, callbackError, callbackDone);
        }
        this.$callback = this.bindCallback(callback);
        this.$callbackDone = this.bindCallback(callbackDone || defaultCallbackDone);
        this.$callbackError = this.bindCallback(callbackError || defaultCallbackError);

        if (callbackDone !== undefined) {
            this.$callbackDone = callbackDone.bind(this);
        }

        if (callbackError !== undefined) {
            this.$callbackError = callbackError.bind(this);
        }

        this.$input = new Rx.Subject();
        this.bindInputs();
        this.$input.subscribe(this.$inputData, this.$inputError, this.$inputClosed);
        this.$output = new Rx.Subject();
        this.output = new Output(this.$output);
    }

    Process.prototype.bindInputs = function() {
        this.$inputData = inputData.bind(this);
        this.$inputError = inputError.bind(this);
        this.$inputClosed = inputClosed.bind(this);
    };

    Process.prototype.$exec = function(input) {
        if (input.Closed) {
            if (input.Error !== undefined) {
                this.$callbackError(input.Error);
            } else {
                this.$callbackDone();
            }
        } else {
            this.$callback(input.Data);
        }
    };

    // will use input as stdin for process
    Process.prototype.connectToStdin = function(input) {
        connectPipes(input, this.$input);
    };

    // will use output as stdout for process
    Process.prototype.connectToStdout = function(output) {
        connectPipes(this.$output, output);
    };

    Process.prototype.bindCallback = function(cb) {
        return cb.bind(this);
    }

    Process.prototype.pipe = function(other, cberr, cbdone) {
        if (!(other instanceof Process)) {
            other = new Process(other, cberr, cbdone);
        }
        this.connectStdout(other.$input);
    };

    Process.pipeline = function(procs) {
        if (procs.length === 0) {
            return undefined;
        }
        var pipeline = new Pipeline();
        var sz = procs.length;
        for (var i = 0; i < sz; i++) {
            pipeline.add(procs[i]);
        }
        return pipeline;
    };

    function connectPipes(source, dest) {
        source.subscribe(function(val){dest.onNext(val);},
            function(err){dest.onError(err);},
            function(){dest.onCompleted();});
    }

    function inputData(data) {
        this.$exec({
            Closed: false,
            Data: data,
        });
    }

    function inputError(err) {
        this.$exec({
            Closed: true,
            Error: err,
        });
    }

    function inputClosed() {
        this.$exec({
            Closed: true,
        });
    }

    E.Process = E.Proc = Process;
}());
