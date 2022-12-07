import * as React from 'react';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import AppBar from '../src/AppBar';
import Copyright from '../src/Copyright';
import PKICard from '../src/PKICard';

export default function Pki() {
  const [pki, setPki] = React.useState([]);
  React.useEffect(() => {
    const run = async () => {
      const resp = await fetch('/config/apps/pki/certificate_authorities');
      const data = await resp.json();
      const caID = Object.keys(data);
      const ca = [];
      for (let i = 0; i < caID.length; i += 1) {
        const id = caID[i];
        const resp = await fetch(`/pki/ca/${id}`);
        const data = await resp.json();
        ca.push(data);
      }
      setPki(ca);
    };
    run();
  }, []);

  return (
    <>
      <AppBar />
      <Container maxWidth="lg">
        <Box sx={{ my: 4 }}>
          <Typography variant="h4" component="h1" gutterBottom>
            Caddy Server PKI
          </Typography>
          <Typography variant="p" component="p" gutterBottom>
            {pki.map((p, i) => {
              return <PKICard key={`n-${i}`} {...p} />;
            })}
          </Typography>
          <Copyright />
        </Box>
      </Container>
    </>
  );
}
