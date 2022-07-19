import * as React from 'react';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import AppBar from '../src/AppBar';
import Copyright from '../src/Copyright';
import UpstreamCard from '../src/UpstreamCard';

export default function Upstream() {
  const [upstream, setUpstream] = React.useState([]);
  React.useEffect(() => {
    const runU = async () => {
      const resp = await fetch('/reverse_proxy/upstreams');
      const data = await resp.json();
      setUpstream(data);
    };
    runU();
  }, []);

  return (
    <>
      <AppBar />
      <Container maxWidth="lg">
        <Box sx={{ my: 4 }}>
          <Typography variant="h4" component="h1" gutterBottom>
            Caddy Server Upstream
          </Typography>
          <Typography variant="p" component="p" gutterBottom>
            {upstream.map((u) => {
              return <UpstreamCard {...u} />;
            })}
          </Typography>
          <Copyright />
        </Box>
      </Container>
    </>
  );
}
