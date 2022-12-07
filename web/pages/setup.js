import * as React from 'react';
import dynamic from 'next/dynamic';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import AppBar from '../src/AppBar';
import Copyright from '../src/Copyright';
const Editor = dynamic(
  () => import("../src/AceEditor"),
  { ssr: false }
)

export default function Load() {
  const [config, setConfig] = React.useState('');
  const [msg, setMsg] = React.useState('');

  const run = React.useCallback(async () => {
    const resp = await fetch('/config/');
    const data = await resp.json();
    setConfig(JSON.stringify(data, null, 2));
    setMsg('');
  }, []);
  React.useEffect(() => {
    run();
  }, []);

  const handleChange = React.useCallback((val) => {
    setConfig(val);
  }, []);

  const save = React.useCallback(
    async (evt) => {
      const r = new Request('/load', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: config,
      });
      const resp = await fetch(r);
      await resp.text();
      setMsg('Save success!');
    },
    [config],
  );

  return (
    <>
      <AppBar />
      <Container maxWidth="lg">
        <Box sx={{ my: 4 }}>
          <Typography variant="h4" component="h1" gutterBottom>
            Caddy Server Load
          </Typography>
          <Typography variant="p" component="p" gutterBottom>
            <Button variant="outlined" onClick={run}>
              Reset
            </Button>
            {'    '}
            <Button variant="contained" onClick={save}>
              Save
            </Button>
          </Typography>
          <Typography variant="p" component="p" gutterBottom style={{minHeight: 600}}>
            <Editor
              onChange={handleChange}
              value={config}
              height={600}
            />
          </Typography>
          <Typography variant="p" component="p" gutterBottom>
            {msg}
          </Typography>
          <Copyright />
        </Box>
      </Container>
    </>
  );
}
