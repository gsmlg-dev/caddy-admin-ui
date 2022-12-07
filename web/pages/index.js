import * as React from 'react';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Accordion from '@mui/material/Accordion';
import AccordionDetails from '@mui/material/AccordionDetails';
import AccordionSummary from '@mui/material/AccordionSummary';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import AppBar from '../src/AppBar';
import Copyright from '../src/Copyright';

export default function Index() {
  const [config, setConfig] = React.useState({});
  React.useEffect(() => {
    const run = async () => {
      const resp = await fetch('/config/');
      const data = await resp.json();
      setConfig(data);
    };
    run();
  }, []);

  const [expanded, setExpanded] = React.useState({});

  const handleChange = (panel) => (event, isExpanded) => {
    setExpanded({
      ...expanded,
      [panel]: isExpanded,
    });
  };

  return (
    <>
      <AppBar />
      <Container maxWidth="lg">
        <Box sx={{ my: 4 }}>
          <Typography variant="h4" component="h1" gutterBottom>
            Caddy Server Config
          </Typography>
          <Typography variant="p" component="pre" gutterBottom>
            {Object.keys(config).map((key) => {
              const subConfig = config[key];
              return (
                <Accordion key={key} expanded={expanded[key]} onChange={handleChange(key)}>
                  <AccordionSummary
                    expandIcon={<ExpandMoreIcon />}
                    aria-controls={`${key}-content`}
                    id={`as-${key}-header`}
                  >
                    <Typography sx={{ width: '33%', flexShrink: 0 }}>{key}</Typography>
                  </AccordionSummary>
                  <AccordionDetails>
                    {key === 'apps' ? (
                      Object.keys(subConfig).map((k) => {
                        const subAppsConfig = subConfig[k];
                        return (
                          <Accordion
                            expanded={expanded[`${key}-${k}`]}
                            onChange={handleChange(`${key}-${k}`)}
                          >
                            <AccordionSummary
                              expandIcon={<ExpandMoreIcon />}
                              aria-controls={`${k}-content`}
                              id={`as-${k}-header`}
                            >
                              <Typography sx={{ width: '33%', flexShrink: 0 }}>
                                {k}
                              </Typography>
                            </AccordionSummary>
                            <AccordionDetails>
                              <Typography component={'pre'}>
                                {JSON.stringify(subAppsConfig, null, 4)}
                              </Typography>
                            </AccordionDetails>
                          </Accordion>
                        );
                      })
                    ) : (
                      <Typography component={'pre'}>
                        {JSON.stringify(subConfig, null, 4)}
                      </Typography>
                    )}
                  </AccordionDetails>
                </Accordion>
              );
            })}
          </Typography>
          <Copyright />
        </Box>
      </Container>
    </>
  );
}
