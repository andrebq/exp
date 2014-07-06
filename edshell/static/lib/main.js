$(function(){
    var EdShell = {};
    EdShell.ActiveEditor = {};
    EdShell.FocusedEditor = {};

    var cmds = new E.CommandMap();
    var db = new E.ShellDB();
    var fs = new E.Fs();

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

    cmds.addCommand("core/save", function(){
        if (!!EdShell.ActiveEditor) {
            var fileName = EdShell.ActiveEditor.getData({ currentFile: "."}).currentFile;
            fs.write(fileName, EdShell.ActiveEditor.getValue())
                .then(function(){
                    EdShell.ActiveEditor.markClean();
                }, function(err){
                    console.log(err);
                    alert('error saving file: ' + err);
                });
        }
    }, EdShell);

    cmds.addCommand("core/reload", function(){
        if (!!EdShell.ActiveEditor) {
            fs.read(EdShell.ActiveEditor.getData({ currentFile: "."}).currentFile).then(function(fileContents){
                EdShell.ActiveEditor.setValue(fileContents);
            });
        }
    }, EdShell);

    cmds.addCommand("core/openFile", function(){
        if (!!EdShell.ActiveEditor) {
            E.Dialog.prompt("which file?").then(function(filename){
                fs.read(filename).then(function(fileContents){
                    EdShell.ActiveEditor.setValue(fileContents);
                    EdShell.ActiveEditor.getData().currentFile = filename;
                });
            });
        }
    }, EdShell);

    function createEditor(holder) {
        var editor = CodeMirror(holder);
        var data = {};
        editor.setData = function(update) {
           data = _.extend(update, data); 
        };

        editor.getData = function(def) {
            if (_(def).isObject()) {
                data = _.extend(def, data);
            }
            return data;
        };

        editor.on("focus", function(cm){
            EdShell.ActiveEditor = cm;
            EdShell.FocusedEditor = cm;
        });

        editor.on("blur", function(cm){
            if (EdShell.FocusedEditor !== cm) {
                EdShell.FocusedEditor = null;
            }
        });
        return editor;
    };

    (function(){
        var initialSize = sizeOf(document.getElementById("content"));
        var mainEditor = createEditor(document.getElementById("content"));
        mainEditor.focus();
        EdShell.mainEditor = mainEditor;
        fullSize(mainEditor);

        document.getElementById("topbar").addEventListener('click', cmds.handler, false);
        $(window).on('resize', function() { fullSize(mainEditor) });
    }());
});
