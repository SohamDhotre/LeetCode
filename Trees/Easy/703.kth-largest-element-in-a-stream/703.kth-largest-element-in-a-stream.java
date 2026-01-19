class KthLargest {

    class MinHeap{
        int []heap;
        int size;
        int cap;

        MinHeap(int cap){
            this.cap=cap;
            this.size=0;
            this.heap=new int[cap];
        }

        int parent(int i){
            return (i-1)/2;
        }

        int leftChild(int i){
            return (2*i)+1;
        }

        int rightChild(int i){
            return (2*i)+2;
        }

        int getMin(){
            if(size==0) return -1;
            else return heap[0];
        }

        void add(int val){
            heap[size]=val;
            int cur=size;
            size++;
            while(cur>0){                
                int parentIndex=parent(cur);
                if (heap[parentIndex] <= heap[cur]) {
                    break;
                }
                int temp=heap[parentIndex];
                heap[parentIndex]=heap[cur];
                heap[cur]=temp;

                cur=parentIndex;
            }
        }

        void removeMin(){
            if (size == 0) {
                return;
            }
            if(size==1){
                size--;
                return;
            }
            heap[0]=heap[size-1];
            size--;
            int cur=0;
            while(true){
                int left=leftChild(cur);
                int right=rightChild(cur);
                int smallest=cur;

                if(left<size && heap[left]<heap[smallest]){
                   smallest=left;
                }
                if(right<size && heap[right]<heap[smallest]){
                   smallest=right;
                }

                if (smallest == cur) {
                    break;
                }

                int temp=heap[cur];
                heap[cur]=heap[smallest];
                heap[smallest]=temp;

                cur=smallest;
            }
        }
    }

    MinHeap heap;

    public KthLargest(int k, int[] nums) {
        heap=new MinHeap(k);
        for(int i:nums) add(i);
    }
    
    public int add(int val) {
        if(heap.size==heap.cap){
            int min=heap.getMin();
            if(val<min) return min;
            else{
                heap.removeMin();
                heap.add(val);
            }
        }
        else heap.add(val);

        return heap.getMin();
    }

}

/**
 * Your KthLargest object will be instantiated and called as such:
 * KthLargest obj = new KthLargest(k, nums);
 * int param_1 = obj.add(val);
 */