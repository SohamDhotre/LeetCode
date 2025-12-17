class MyStack {
    int size;
    Queue<Integer>q;
    public MyStack() {
        size=0;
        q=new ArrayDeque<>();
    }
    
    public void push(int x) {
        q.add(x);
        size++;
        if(size==1) return;
        int count=size;
        // System.out.println("size: "+size);
        while(count>1){
            // System.out.println("q: "+q);
            // System.out.println(q.peek());
            q.add(q.remove());
            count--;
        }
    }
    
    public int pop() {
        size--;
        return q.remove();
    }
    
    public int top() {
        // System.out.println("from top q: "+q);
        return q.peek();
    }
    
    public boolean empty() {
        return size==0;
    }
}

/**
 * Your MyStack object will be instantiated and called as such:
 * MyStack obj = new MyStack();
 * obj.push(x);
 * int param_2 = obj.pop();
 * int param_3 = obj.top();
 * boolean param_4 = obj.empty();
 */