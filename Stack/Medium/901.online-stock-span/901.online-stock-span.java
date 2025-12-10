class StockSpanner {
    ArrayDeque<int[]>st;
    int day;
    public StockSpanner() {
        st=new ArrayDeque<>();
        day=0;
    }
    
    public int next(int price) {
        day++;
        // System.out.println("day: "+day);
        while(!st.isEmpty() && st.peek()[0]<=price) {
            st.pop();
        }
        // System.out.println("before push st: ");
        // st.forEach(k->System.out.print(Arrays.toString(k)));
        // System.out.println("st.peek(): "+Arrays.toString(st.peek()));
        int span = st.isEmpty() ? day : day - st.peek()[1];
        st.push(new int[]{price, day});
        // System.out.println("day-st.peek()[1]: "+(day-st.peek()[1]));
        return span;
    }
}

/**
 * Your StockSpanner object will be instantiated and called as such:
 * StockSpanner obj = new StockSpanner();
 * int param_1 = obj.next(price);
 */