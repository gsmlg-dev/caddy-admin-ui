import * as React from 'react';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import AppBar from '../src/AppBar';
import Copyright from '../src/Copyright';

export default function Load() {
  React.useEffect(()=> {
    const run = async () => {
    };
    run();
  }, []);

  return (
    <>
      <AppBar />
      <Container maxWidth="lg">
        <Box sx={{ my: 4 }}>
          <Typography variant="h4" component="h1" gutterBottom>
            Caddy Server Load
          </Typography>
          <Typography variant="p" component="pre" gutterBottom>
          </Typography>
          <Copyright />
        </Box>
      </Container>
    </>
  );
}
