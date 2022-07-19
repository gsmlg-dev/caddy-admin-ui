import * as React from 'react';
import Container from '@mui/material/Container';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import AppBar from '../src/AppBar';
import OutlinedInput from '@mui/material/OutlinedInput';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import Select from '@mui/material/Select';
import Chip from '@mui/material/Chip';
import { TheatersRounded } from '@mui/icons-material';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';

const parseMetrics = (txt) => {
  const lines = txt.split(/\n|\r\n/);
  const data = {};
  lines.forEach((line) => {
    if (/(^\s+)?#/.test(line)) {
      return;
    }
    const [key, value] = line.split(/\s+/);
    if (!key) return;
    data[key] = Number(value);
  });
  return data;
};

function TabPanel(props) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`vertical-tabpanel-${index}`}
      aria-labelledby={`vertical-tab-${index}`}
      {...other}
    >
      {value === index && (
        <Box sx={{ p: 3 }}>
          <Typography>{children}</Typography>
        </Box>
      )}
    </div>
  );
}

function a11yProps(index) {
  return {
    id: `vertical-tab-${index}`,
    'aria-controls': `vertical-tabpanel-${index}`,
  };
}

const ITEM_HEIGHT = 48;
const ITEM_PADDING_TOP = 8;
const MenuProps = {
  PaperProps: {
    style: {
      maxHeight: ITEM_HEIGHT * 4.5 + ITEM_PADDING_TOP,
      width: 1000,
    },
  },
};

function getSelectedStyles(name, pName) {
  return {
    fontWeight: pName.indexOf(name) === -1 ? 'normal' : '500',
  };
}

export default function Monitor() {
  const [rawMetrics, setRawMetrics] = React.useState('');
  const [metrics, setMetrics] = React.useState([]);
  const [value, setValue] = React.useState(0);

  const handleChange = (event, newValue) => {
    setValue(newValue);
  };

  const run = React.useCallback(async () => {
    const resp = await fetch('/metrics');
    const data = await resp.text();
    setRawMetrics(data);
    const parsed = parseMetrics(data);
    const n = [{ time: Date.now(), data: parsed }].concat(metrics).slice(0, 240);
    setMetrics(n);
  }, [metrics]);
  React.useEffect(() => {
    run();
    const t = setInterval(() => {
      run();
    }, 30_000);
    return () => clearInterval(t);
  }, []);

  const [selectedMetrics, setSelectedMetrics] = React.useState([]);

  const handleSelectChange = (event) => {
    const {
      target: { value },
    } = event;
    setSelectedMetrics(
      // On autofill we get a stringified value.
      typeof value === 'string' ? value.split(',') : value,
    );
  };

  const latestMetric = metrics[0];
  const metricNames = Object.keys(latestMetric?.data ?? {});

  return (
    <>
      <AppBar />
      <Container maxWidth="lg">
        <Box sx={{ my: 4 }}>
          <Typography variant="h4" component="h1" gutterBottom>
            Caddy Server Metrics
          </Typography>
          <Box
            sx={{
              flexGrow: 1,
              bgcolor: 'background.paper',
              display: 'flex',
              height: 224,
            }}
          >
            <Tabs
              orientation="vertical"
              variant="scrollable"
              value={value}
              onChange={handleChange}
              aria-label="Vertical tabs example"
              sx={{ borderRight: 1, borderColor: 'divider', minWidth: 100 }}
            >
              <Tab label="Raw" {...a11yProps(0)} />
              <Tab label="Parsed" {...a11yProps(1)} />
              <Tab label="Metric" {...a11yProps(2)} />
            </Tabs>
            <TabPanel value={value} index={0}>
              <pre>{rawMetrics}</pre>
            </TabPanel>
            <TabPanel value={value} index={1}>
              <pre>{JSON.stringify(metrics[0], null, 4)}</pre>
            </TabPanel>
            <TabPanel value={value} index={2}>
              <div>
                <FormControl sx={{ m: 1, width: 800 }}>
                  <InputLabel id="demo-multiple-chip-label">Select Metric</InputLabel>
                  <Select
                    labelId="demo-multiple-chip-label"
                    id="demo-multiple-chip"
                    multiple
                    value={selectedMetrics}
                    onChange={handleSelectChange}
                    input={<OutlinedInput id="select-multiple-chip" label="Chip" />}
                    renderValue={(selected) => (
                      <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                        {selected.map((value) => (
                          <Chip key={value} label={value} />
                        ))}
                      </Box>
                    )}
                    MenuProps={MenuProps}
                  >
                    {metricNames.map((name) => (
                      <MenuItem
                        key={name}
                        value={name}
                        style={getSelectedStyles(name, selectedMetrics)}
                      >
                        {name}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </div>
              <Table sx={{ minWidth: 800 }} aria-label="simple table">
                <TableHead>
                  <TableRow>
                    <TableCell>time</TableCell>
                    {selectedMetrics.map((m) => (
                      <TableCell key={`m-${m}`}>{m}</TableCell>
                    ))}
                  </TableRow>
                </TableHead>
                <TableBody>
                  {metrics.map((m) => {
                    return (
                      <TableRow key={`t-${m.time}`}>
                        <TableCell>{new Date(m.time).toISOString()}</TableCell>
                        {selectedMetrics.map((sm) => (
                          <TableCell key={`t-${m.time}-${sm}`}>{m.data[sm]}</TableCell>
                        ))}
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </TabPanel>
          </Box>
        </Box>
      </Container>
    </>
  );
}
