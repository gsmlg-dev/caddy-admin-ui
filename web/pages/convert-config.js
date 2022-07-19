import * as React from 'react';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import TextareaAutosize from '@mui/material/TextareaAutosize';
import Editor from 'react-simple-code-editor';
import { highlight, languages } from 'prismjs/components/prism-core';
import 'prismjs/components/prism-clike';
import 'prismjs/components/prism-log';
import 'prismjs/themes/prism.css'; //Example style, you can use another
import Button from '@mui/material/Button';
import AppBar from '../src/AppBar';
import Copyright from '../src/Copyright';

export default function ConvertConfig() {
  const [config, setConfig] = React.useState('');
  const [parsedConfig, setParsedConfig] = React.useState();

  const handleChange = React.useCallback((val) => {
    setConfig(val);
  }, []);

  const handleHighlight = React.useCallback((code) => highlight(code, languages.log), []);

  const save = React.useCallback(async (evt) => {
    const r = new Request('/adapt', {
      method: 'POST',
      headers: {
        'Content-Type': 'text/caddyfile',
      },
      body: config,
    });
    const resp = await fetch(r);
    const convertData = await resp.json();
    console.log(convertData);
    setParsedConfig(convertData);
  }, [config]);

  return (
    <>
      <AppBar />
      <Container maxWidth="lg">
        <Box sx={{ my: 4 }}>
          <Typography variant="h4" component="h1" gutterBottom>
            Caddy Server Load
          </Typography>
          <Typography variant="p" component="p" gutterBottom>
            <Button variant="contained" onClick={save}>Convert</Button>
          </Typography>
          <Typography variant="p" component="p" gutterBottom>
            <Editor
              onValueChange={handleChange}
              value={config}
              highlight={handleHighlight}
              padding={10}
              style={{
                fontFamily: '"Fira code", "Fira Mono", monospace',
                fontSize: 16,
                border: '1px solid #aaa',
              }}
            />
          </Typography>
          <Typography variant="p" component="p" gutterBottom>
            <TextareaAutosize
              style={{ width: '100%' }}
              value={JSON.stringify(parsedConfig, null, 2)}
            />
          </Typography>
          <Copyright />
        </Box>
      </Container>
    </>
  );
}
