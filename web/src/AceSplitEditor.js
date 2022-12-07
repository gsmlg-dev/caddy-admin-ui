import React from "react";
import { split as SplitEditor } from "react-ace";
import { ParentSize } from '@gsmlg/visx';
import "ace-builds/src-noconflict/mode-json";
import "ace-builds/src-noconflict/theme-dracula";
import "ace-builds/src-noconflict/ext-language_tools";

const Editor = (props) => {
  console.log(props);
  return (
    <ParentSize>
      {({ width, height }) => (
        <SplitEditor
          width={width}
          height={height}
          mode="json"
          theme="dracula"
          splits={2}
          orientation="below"
          name="UNIQUE_ID_OF_DIV"
          editorProps={{ $blockScrolling: true }}
          {...props}
        />
      )}
    </ParentSize>
  )
};

export default Editor;
