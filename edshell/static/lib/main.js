$(function(){
    function CommandMap(opts) {
        if (!(this instanceof CommandMap)) {
            return new CommandMap();
        }
        this.$constructor(opts);
    }

    CommandMap.prototype.$constructor = function(opts) {
        opts = opts || {};
        this.$opts = opts;
        this.$cmds = {};
        this.handler = this.$handler.bind(this);
    };

    CommandMap.prototype.$handler = function(ev) {
        this.process(ev.target.getAttribute("data-command"), ev, ev.target);
    };

    CommandMap.prototype.addCommand = function(name, handler) {
        this.$cmds[name] = handler;
    };

    CommandMap.prototype.process = function(name, opt, ctx) {
        if (!!this.$cmds[name]) {
            this.$cmds[name].apply(ctx, [opt]);
        }
    };

    function sizeOf(element) {
        return {
            height: $(element).height(),
            width: $(element).width(),
        };
    }

    function topbarSize() {
        return sizeOf(document.getElementById("topbar"));
    }

    function bodySize() {
        return sizeOf(document.body);
    }

    function contentHeight(editor) {
        var bs = bodySize(),
            tb = topbarSize(),
            ss = sizeOf(editor.getWrapperElement().parentElement);
            ss.height = bs.height - tb.height;
            return ss;
    }

    function fullSize(editor) {
        var holder = editor.getWrapperElement().parentElement;
        sectionSz = contentHeight(editor);
        resizeEditor(editor, sectionSz);
    }

    function resizeEditor(editor, newSize) {
        editor.setSize(newSize.width, newSize.height);
        editor.refresh();
    }

    function saveCurrentFile(ev) {
    }

    function reloadCurrentFile(ev) {
    }

    function loadSession() {
        return ShellDB().fetch("db/session");
    }

    function loadFile(fileName) {
        return Rx.Observable.fromPromise(
            $.get(URI("/fs/" + fileName).normalizePathname()));
    }

    var cmds = new CommandMap();
    var db = new ShellDB();

    var pipeline = new E.Proc.pipeline([
        function(input){
            console.log("first ", input);
            this.output.write(input);
        },
        function(input){
            console.log("second ", input);
            this.output.write(input);
        },
    ]);

    (function(){
        var initialSize = sizeOf(document.getElementById("content"));
        var mainContent = CodeMirror(document.getElementById("content"));
        fullSize(mainContent);
        cmds.addCommand("core/save", saveCurrentFile);
        cmds.addCommand("core/reload", reloadCurrentFile);

        pipeline.head.connectToStdin(Rx.Observable
            .interval(500)
            .timeInterval()
            .take(10));

        document.getElementById("topbar").addEventListener('click', cmds.handler, false);
        $(window).on('resize', function() { fullSize(mainContent) });
    }());
});
