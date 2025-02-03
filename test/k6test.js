import http from 'k6/http';
import { check, sleep } from 'k6';

// Настройки теста
export const options = {
    stages: [
        { duration: '1s', target: 100 },  // Постепенно разогреваем до 100 пользователей
        { duration: '30s', target: 2000 }, // Держим 1000 одновременных пользователей
        { duration: '1s', target: 100 },    // Плавное завершение
    ],
};

// Уникальный ID кошелька
const walletID = 'b49c66cb-90f8-4ad7-b2b3-fd993e9d9efd'; // Замените на реальный UUID из вашего приложения
const baseURL = 'http://localhost:8080/api/v1';          // URL вашего приложения

// Основная функция теста
export default function () {
    // Тестируем депозит
    const depositPayload = JSON.stringify({
        walletId: walletID,
        operationType: 'DEPOSIT',
        amount: 10,
    });
    const depositHeaders = { 'Content-Type': 'application/json' };

    const depositResponse = http.post(`${baseURL}/wallet`, depositPayload, { headers: depositHeaders });
    check(depositResponse, {
        'Deposit status is 200': (r) => r.status === 200,
    });

    const withdrawPayload = JSON.stringify({
        walletId: walletID,
        operationType: 'WITHDRAW',
        amount: 5,
    });

    const withdrawHeaders = { 'Content-Type': 'application/json' };

    const withdrawResponse = http.post(`${baseURL}/wallet`, withdrawPayload, { headers: withdrawHeaders });
    check(withdrawResponse, {
        'Withdraw status is 200': (r) => r.status === 200,
    });

    // Тестируем получение баланса
    const balanceResponse = http.get(`${baseURL}/balance/${walletID}`);
    check(balanceResponse, {
        'Get balance status is 200': (r) => r.status === 200,
        'Balance contains amount': (r) => r.json('balance') >= 0,
    });

    // Задержка между запросами для имитации реального поведения
   sleep(0.1);
}