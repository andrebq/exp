(function(){
    var state = {
        createdBuffers: 0,
    };
    (function setupCodeMirror(){
        CodeMirror.commands.save = function(cm) {
            CodeMirror.signal(cm, "editor-save", cm);
        };
    }());
    Polymer('ed-editor', {
        resize: function(newsize) {
            if (this.$cm) {
                this.$doCmResize(newsize);
            }
        },
        isUnamedBuffer: function() {
            return S(this.$bufferName).startsWith("no-name-");
        },
        isClean: function() {
            return this.$cm.isClean();
        },
        markClean: function() {
            this.$cm.markClean();
        },
        getValue: function() {
            return this.$cm.getValue();
        },
        setValue: function(value) {
            this.$cm.setValue(value);
        },
        setBufferName: function(name) {
            this.$bufferName = name;
        },
        getBufferName: function() {
            return this.$bufferName;
        },
        created: function() {
            this.$subs = new E.Rx.Util.SubManager();
            this.$bufferName = "no-name-" + state.createdBuffers;
            state.createdBuffers++;
        },
        attached: function() {
            E.Css.loadStyle(this.resolvePath("../bower_components/codemirror/lib/codemirror.css"), this.shadowRoot)
                .then(this.createEditor.bind(this));
        },
        createEditor: function() {
            var ta = this.ownerDocument.createElement("textarea");
            ta.innerText = "";
            this.shadowRoot.appendChild(ta);
            this.$cm = CodeMirror.fromTextArea(ta, {
                lineNumbers: true,
                viewportMargin: 100,
            });
            var dim = E.Rx.dimension(this);
            this.$doCmResize(dim);
            this.fire("editor-created");
            this.$cm.on("editor-save", function(){
                this.fire("editor-save", {});
            }.bind(this));
        },
        $doCmResize: function(dim) {
            this.$cm.setSize(dim.width + "px", dim.height + "px");
            window.requestAnimationFrame(function(){
                this.$cm.refresh();
            }.bind(this));
        },
        focus: function() {
            this.$cm.focus();
            this.fire("editor-focused");
        },
    });
}.bind(window)());
