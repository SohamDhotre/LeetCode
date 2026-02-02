class MedianFinder {
    PriorityQueue<Integer>leftMaxHeap;
    PriorityQueue<Integer>rightMinHeap;
    public MedianFinder() {
        leftMaxHeap=new PriorityQueue<>(Collections.reverseOrder());
        rightMinHeap=new PriorityQueue<>();
    }
    
    public void addNum(int num) {
        if(leftMaxHeap.size()==0){
            leftMaxHeap.add(num);
        }
        else if(num<=leftMaxHeap.peek()) leftMaxHeap.add(num);
        else rightMinHeap.add(num);
        
        if(Math.abs(leftMaxHeap.size()-rightMinHeap.size())>1){
            if(leftMaxHeap.size()>rightMinHeap.size()) shiftElementFromLeftHeap();
            else shiftElementFromRightHeap();
        }
    }

    public double findMedian() {
        if(leftMaxHeap.size()==rightMinHeap.size())
            return (leftMaxHeap.peek()+rightMinHeap.peek())/2.0;
        else if(leftMaxHeap.size()>rightMinHeap.size()) 
            return leftMaxHeap.peek();
        return rightMinHeap.peek(); 
    }

    void shiftElementFromLeftHeap(){
        rightMinHeap.add(leftMaxHeap.poll());
    }

    void shiftElementFromRightHeap(){
        leftMaxHeap.add(rightMinHeap.poll());
    }
}

/**
 * Your MedianFinder object will be instantiated and called as such:
 * MedianFinder obj = new MedianFinder();
 * obj.addNum(num);
 * double param_2 = obj.findMedian();
 */