function worker() {
  console.log('sleeping for 60 seconds');
  setTimeout(worker, 60 * 1000);
}

worker();
