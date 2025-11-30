/*
// Definition for a Node.
class Node {
    public int val;
    public Node prev;
    public Node next;
    public Node child;
};
*/

class Solution {
    public Node flatten(Node head) {
        Stack<Node>st=new Stack<>();
        Node cur=head, prev=null;
        while(cur!=null || !st.isEmpty()){
            if(cur==null && !st.isEmpty()){ 
                // System.out.println("if condition entered ");   
                // print(head);
                cur=st.pop();
                prev.next=cur;
                cur.prev=prev;
            }
            
            if(cur.child!=null) {
                if(cur.next!=null) st.push(cur.next);
                cur.next=cur.child;
                cur.child=null;
                prev=cur;
                cur=cur.next;
                cur.prev=prev;
            } else {
                prev=cur;
                cur=cur.next;
            }            
        }
        // System.out.print("return list: ");
        // print(head);
        return head;
    }

    void print(Node head){
        Node cur=head, prev=null;
        System.out.print("\n\nlist: ");
        while(cur!=null){
            System.out.print(cur.val+"(child: "+(cur.child!=null?cur.child.val:"null")+") -> ");
            prev=cur;
            cur=cur.next;
        }
        System.out.print(" null \n rev list: ");
        while(prev!=null){
            System.out.print(prev.val+"(child: "+(prev.child!=null?prev.child.val:"null")+") -> ");
            prev=prev.prev;
        }
        System.out.println(" null\n\n");
    }
}