let currentNumber = '';
let previousNumber = '';
let operation = null;

function appendNumber(number) {
  currentNumber += number;
  updateDisplay();
}

function updateDisplay() {
  document.getElementById('display').value = currentNumber;
}

function setOperation(op) {
  if (currentNumber === '') return;
  if (previousNumber !== '') {
    calculate();
  }
  operation = op;
  previousNumber = currentNumber;
  currentNumber = '';
}

function calculate() {
  let result;
  const prev = parseFloat(previousNumber);
  const curr = parseFloat(currentNumber);
  if (isNaN(prev) || isNaN(curr)) return;
  switch (operation) {
    case '+':
      result = prev + curr;
      break;
    case '-':
      result = prev - curr;
      break;
    case '*':
      result = prev * curr;
      break;
    case '/':
      result = prev / curr;
      break;
    default:
      return;
  }
  currentNumber = result.toString();
  operation = null;
  previousNumber = '';
  updateDisplay();
}

function clearDisplay() {
  currentNumber = '';
  previousNumber = '';
  operation = null;
  updateDisplay();
}
