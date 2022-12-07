import * as React from 'react';
import dynamic from 'next/dynamic';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import AppBar from '../src/AppBar';
import Copyright from '../src/Copyright';
const Editor = dynamic(
  () => import("../src/AceSplitEditor"),
  { ssr: false }
)

export default function ConvertConfig() {
  const [config, setConfig] = React.useState(['caddy.json', 'Caddyfile']);

  const handleChange = (val) => {
    setConfig(val);
  };

  const save = React.useCallback(
    async (evt) => {
      const r = new Request('/adapt', {
        method: 'POST',
        headers: {
          'Content-Type': 'text/caddyfile',
        },
        body: config[0],
      });
      const resp = await fetch(r);
      const convertData = await resp.json();

      setConfig([config[0], JSON.stringify(convertData, null, 2)]);
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
            <Button variant="contained" onClick={save}>
              Convert
            </Button>
          </Typography>
          <Typography variant="p" component="p" gutterBottom>
            <Editor
              onChange={console.log}
              value={config}
              height={600}
              splits={2}
            />
          </Typography>
          <Copyright />
        </Box>
      </Container>
    </>
  );
}
