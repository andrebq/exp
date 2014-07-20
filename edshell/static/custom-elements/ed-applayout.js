(function(){
    function setFocus(el) {
        if (_.isFunction(el.focus, 'focus')) {
            el.focus();
        }
    }

    function notifyResize(node, newsize) {
        if (_.isFunction(node.resize)) {
            node.resize(newsize);
        }
    }

    function notifyHide(node) {
        if (_.isFunction(node.hide)) {
            node.hide();
        }
    }

    function notifyShow(node) {
        if (_.isFunction(node.show)) {
            node.show();
        }
    }

    function notNull(e) {
        return e !== null;
    }

    function translate(e, xy) {
        E.Css(e).set("transform",sprintf("translate(%fpx,%fpx)", xy.x, xy.y));
    }

    function removePaddingFromParent(applayout) {
        var parent = applayout.parentElement;
        parent.style.padding = 0;
        parent.style.margin = 0;
    }

    function hideNodes(nodes) {
        _(nodes).filter(notNull).forEach(function(e){
            e.style.display = 'none';
        });
    }

    function showNodes(nodes) {
        _(nodes).filter(notNull).forEach(function(e){
            e.style.display= 'block';
            e.style.position = 'absolute';
            e.style.padding= '0';
            e.style.margin= '0';
        });
    }

    function resizeNodes(nodes, size) {
        var visible = _(nodes).filter(notNull).filter(function(e){ return e.style.display !== 'none'; }).length;
        _(nodes).filter(notNull).filter(function(e){ return e.style.display !== 'none'; })
        .forEach(function(e, idx){
            var width = (size.width / visible);
            var margin = width * idx;
            e.style.width = width + "px";
            e.style.height = (size.height) + "px";
            translate(e, {x: margin, y: 0});
            notifyResize(e, {width: width, height: size.height});
        });
    }

    function resizeSidebar(sidebar, size) {
        if (sidebar === null) { return; }
        size = _.extend(size, { width: size.width * 0.75})
        E.Css(sidebar).setSize(size);
        notifyResize(sidebar, size);
    }

    Polymer('ed-applayout', {
        sidebar: "ed-sidebar",
        right: "#right",
        middle: "#middle",
        left: "#left",
        mode: "single",
        saveLastFocusedEditor: function(target) {
            this.$lastFocusedEditor = target;
        },
        attached: function() {
            this.$subs.add("editor-focused", Rx.Observable.fromEvent(this, 'editor-focused')
                .pluck('target')
                .filter(E.Rx.distinctFromLast())
                .subscribe(this.saveLastFocusedEditor.bind(this)));
            this.$subs.add("editor-created", Rx.Observable.fromEvent(this, 'editor-created')
                .pluck('target')
                .subscribe(setFocus));
            this.$subs.add("window-f2", Rx.Observable.fromEvent(window, 'keyup')
                .pluck('which')
                .filter(E.Rx.isKeyCode(E.Rx.Keycodes.F2))
                .subscribe(function(key){
                    this.toggleSideBar(true);
                }.bind(this)));
            this.$subs.add("search-completed", Rx.Observable.fromEvent(this, "search-completed")
                .pluck('detail').pluck('value')
                .subscribe(function(value){
                    this.toggleSideBar(false);
                }.bind(this)));
            this.$subs.add("window-resize", Rx.Observable
                .fromEvent(window, 'resize')
                .pluck("target")
                .map(E.Rx.dimension)
                .filter(E.Rx.distinctFromLast())
                .subscribe(this.handleResize.bind(this)));
            var nodes = this.$getNodes();
            // the side bar start hidden
            hideNodes([nodes.sidebar]);
            // schedule a resize
            setTimeout(function(){
                this.handleResize(E.Rx.dimension(window));
            }.bind(this), 1);
            removePaddingFromParent(this);
        },
        handleResize: function(newsize) {
            if (newsize === undefined) {
                newsize = this.lastsize;
            }
            var nodes = this.$getNodes();
            var mode = this.mode;
            if (mode === "single") {
                hideNodes([nodes.middle, nodes.right]);
                showNodes([nodes.left]);
            } else if (mode === "dual") {
                showNodes([nodes.left, nodes.middle]);
            } else if (mode === "all") {
                showNodes([nodes.left, nodes.middle, nodes.right]);
            }
            resizeNodes([this], newsize);
            resizeNodes([nodes.left, nodes.middle, nodes.right], newsize);
            resizeSidebar(nodes.sidebar, newsize);
            this.lastsize = newsize;
        },
        detached: function() {
            this.$subs.dispose("window-resize");
        },
        created: function() {
            this.$subs = new E.Rx.Util.SubManager();
        },
        attributeChanged: function(name, oldval, newval) {
            if (name === "mode") {
                this.handleResize();
            }
        },
        $getNodes: function() {
            return {
                sidebar: this.querySelector(this.sidebar),
                left: this.querySelector(this.left),
                right: this.querySelector(this.right),
                middle: this.querySelector(this.middle),
            }
        },
        toggleSideBar: function(show) {
            var nodes = this.$getNodes();
            if (nodes.sidebar) {
                if (show) {
                    E.Css(nodes.sidebar).set("z-index", 9999);
                    showNodes([nodes.sidebar]);
                    notifyShow(nodes.sidebar);
                } else {
                    hideNodes([nodes.sidebar]);
                    notifyHide(nodes.sidebar);
                    if (this.$lastFocusedEditor) {
                        this.$lastFocusedEditor.focus();
                    }
                }
            }
        },
    });
}.bind(window)());
