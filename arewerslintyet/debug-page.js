const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: false });
  const page = await browser.newPage();
  
  try {
    console.log('Navigating to http://localhost:3001');
    await page.goto('http://localhost:3001');
    
    console.log('Waiting for page to load...');
    await page.waitForLoadState('networkidle');
    
    console.log('Getting page content...');
    const content = await page.content();
    console.log('Page HTML length