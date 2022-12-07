import React from "react";
import AceEditor from "react-ace";
import { ParentSize } from '@gsmlg/visx';
import "ace-builds/src-noconflict/mode-json";
import "ace-builds/src-noconflict/theme-dracula";
import "ace-builds/src-noconflict/ext-language_tools";

function onChange(newValue) {
  console.log("change", newValue);
}

const Editor = (props) => {
  return (
    <ParentSize>
      {({ width, height }) => (
        <AceEditor
          width={width}
          height={height}
          mode="json"
          theme="dracula"
          onChange={onChange}
          editorProps={{ $blockScrolling: true }}
          {...props}
        />
      )}
    </ParentSize>
)
};

export default Editor;
